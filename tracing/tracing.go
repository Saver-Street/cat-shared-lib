// Package tracing provides helpers for initialising and using OpenTelemetry
// distributed tracing. It wraps the OTel SDK with sensible defaults and
// exposes a small, ergonomic API for setting up a TracerProvider, creating
// spans, and propagating trace context over HTTP.
//
// Usage:
//
//	tp, err := tracing.NewProvider(ctx, tracing.Config{
//	    ServiceName:    "my-service",
//	    ServiceVersion: "v1.2.3",
//	    Exporter:       tracing.ExporterStdout,
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer tp.Shutdown(ctx)
//
//	tracer := tp.Tracer("component-name")
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
package tracing

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
	"go.opentelemetry.io/otel/trace"
)

// ExporterType selects which span exporter to use.
type ExporterType string

const (
	// ExporterStdout writes human-readable spans to stdout.
	ExporterStdout ExporterType = "stdout"
	// ExporterNoop discards all spans. Useful in tests.
	ExporterNoop ExporterType = "noop"
)

// Config configures the TracerProvider.
type Config struct {
	// ServiceName is the logical name of the service (required).
	ServiceName string
	// ServiceVersion is the version string (e.g. "v1.0.0").
	ServiceVersion string
	// Environment is the deployment environment (e.g. "production").
	Environment string
	// Exporter selects the span exporter. Default: ExporterNoop.
	Exporter ExporterType
	// SampleRate is the fraction of traces to sample [0.0, 1.0]. Default: 1.0.
	SampleRate float64
}

func (c *Config) defaults() {
	if c.Exporter == "" {
		c.Exporter = ExporterNoop
	}
	if c.SampleRate == 0 {
		c.SampleRate = 1.0
	}
}

// Provider wraps an OTel TracerProvider with a Shutdown method.
type Provider struct {
	tp *sdktrace.TracerProvider
}

// NewProvider creates and configures an OTel TracerProvider. Call Shutdown
// when the application exits to flush remaining spans.
func NewProvider(ctx context.Context, cfg Config) (*Provider, error) {
	cfg.defaults()

	exp, err := newExporter(ctx, cfg.Exporter)
	if err != nil {
		return nil, fmt.Errorf("tracing: create exporter: %w", err)
	}

	res, err := newResource(cfg)
	if err != nil {
		return nil, fmt.Errorf("tracing: create resource: %w", err)
	}

	sampler := sdktrace.TraceIDRatioBased(cfg.SampleRate)
	if cfg.SampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Register as global provider and set composite propagator.
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Provider{tp: tp}, nil
}

// Tracer returns a Tracer from the provider with the given instrumentation name.
func (p *Provider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return p.tp.Tracer(name, opts...)
}

// Shutdown flushes and stops the provider. It should be deferred after New.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p == nil || p.tp == nil {
		return nil
	}
	if err := p.tp.Shutdown(ctx); err != nil {
		return fmt.Errorf("tracing: shutdown: %w", err)
	}
	return nil
}

// ---- Span helpers ----

// Start starts a new span and returns the updated context and the span.
// The caller is responsible for calling span.End().
func Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return otel.Tracer("").Start(ctx, name, opts...)
}

// StartWithTracer starts a span using the given tracer.
func StartWithTracer(ctx context.Context, t trace.Tracer, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.Start(ctx, name, opts...)
}

// RecordError records an error on the span and sets its status to Error.
// It is a no-op if err is nil.
func RecordError(span trace.Span, err error, opts ...trace.EventOption) {
	if err == nil {
		return
	}
	span.RecordError(err, opts...)
	span.SetStatus(codes.Error, err.Error())
}

// SetAttributes sets key-value attributes on the span.
func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// TraceID returns the trace ID from the context as a hex string, or an empty
// string if the context carries no valid span.
func TraceID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// SpanID returns the span ID from the context as a hex string, or an empty
// string if the context carries no valid span.
func SpanID(ctx context.Context) string {
	sc := trace.SpanFromContext(ctx).SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.SpanID().String()
}

// IsRecording reports whether the span in ctx is currently recording.
func IsRecording(ctx context.Context) bool {
	return trace.SpanFromContext(ctx).IsRecording()
}

// ---- HTTP middleware ----

// Middleware returns an http.Handler that extracts trace context from incoming
// requests and starts a server span for each request.
func Middleware(tracer trace.Tracer, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := otel.GetTextMapPropagator().Extract(r.Context(), propagation.HeaderCarrier(r.Header))
		spanName := r.Method + " " + r.URL.Path
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(r.Method),
				semconv.URLPath(r.URL.Path),
			),
		)
		defer span.End()

		rw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r.WithContext(ctx))

		span.SetAttributes(semconv.HTTPResponseStatusCode(rw.status))
		if rw.status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(rw.status))
		}
	})
}

// InjectHTTP injects trace context from ctx into the outgoing HTTP request
// headers. Use when making downstream HTTP calls.
func InjectHTTP(ctx context.Context, req *http.Request) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(req.Header))
}

// ExtractHTTP extracts trace context from incoming HTTP request headers into
// the returned context.
func ExtractHTTP(ctx context.Context, req *http.Request) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.HeaderCarrier(req.Header))
}

// statusWriter wraps ResponseWriter to capture the status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

// ---- internal helpers ----

func newExporter(ctx context.Context, t ExporterType) (sdktrace.SpanExporter, error) {
	switch t {
	case ExporterStdout:
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case ExporterNoop:
		return &noopExporter{}, nil
	default:
		return nil, fmt.Errorf("tracing: unknown exporter type %q", t)
	}
}

func newResource(cfg Config) (*resource.Resource, error) {
	attrs := []attribute.KeyValue{
		semconv.ServiceNameKey.String(cfg.ServiceName),
	}
	if cfg.ServiceVersion != "" {
		attrs = append(attrs, semconv.ServiceVersionKey.String(cfg.ServiceVersion))
	}
	if cfg.Environment != "" {
		attrs = append(attrs, semconv.DeploymentEnvironmentNameKey.String(cfg.Environment))
	}
	return resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
}

// noopExporter discards all spans.
type noopExporter struct{}

func (n *noopExporter) ExportSpans(_ context.Context, _ []sdktrace.ReadOnlySpan) error {
	return nil
}
func (n *noopExporter) Shutdown(_ context.Context) error { return nil }
