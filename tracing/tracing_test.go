package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func newTestProvider(t *testing.T) *Provider {
	t.Helper()
	tp, err := NewProvider(context.Background(), Config{
		ServiceName:    "test-svc",
		ServiceVersion: "v0.0.1",
		Exporter:       ExporterNoop,
	})
	testkit.RequireNoError(t, err)
	t.Cleanup(func() { tp.Shutdown(context.Background()) })
	return tp
}

func TestNewProvider_Noop(t *testing.T) {
	tp := newTestProvider(t)
	testkit.AssertNotNil(t, tp)
}

func TestNewProvider_Stdout(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "test",
		Exporter:    ExporterStdout,
	})
	testkit.RequireNoError(t, err)
	tp.Shutdown(context.Background())
}

func TestNewProvider_UnknownExporter(t *testing.T) {
	_, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    "kafka",
	})
	testkit.AssertError(t, err)
}

func TestNewProvider_WithEnvironment(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Environment: "production",
		Exporter:    ExporterNoop,
	})
	testkit.RequireNoError(t, err)
	tp.Shutdown(context.Background())
}

func TestNewProvider_SampleRateZeroDefaults(t *testing.T) {
	cfg := Config{ServiceName: "s", SampleRate: 0}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.SampleRate, 1.0)
}

func TestNewProvider_ExporterDefault(t *testing.T) {
	cfg := Config{ServiceName: "s"}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.Exporter, ExporterNoop)
}

func TestProvider_Tracer(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("test-component")
	testkit.AssertNotNil(t, tr)
}

func TestProvider_Shutdown(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    ExporterNoop,
	})
	testkit.RequireNoError(t, err)
	testkit.AssertNoError(t, tp.Shutdown(context.Background()))
}

func TestNewProvider_RatioSampler(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    ExporterNoop,
		SampleRate:  0.5,
	})
	testkit.RequireNoError(t, err)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	testkit.AssertNotNil(t, tp)
}

func TestNewProvider_FullSampler(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    ExporterNoop,
		SampleRate:  1.0,
	})
	testkit.RequireNoError(t, err)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	testkit.AssertNotNil(t, tp)
}

func TestNewProvider_WithVersion(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName:    "svc",
		Exporter:       ExporterNoop,
		ServiceVersion: "1.2.3",
		Environment:    "test",
	})
	testkit.RequireNoError(t, err)
	t.Cleanup(func() { _ = tp.Shutdown(context.Background()) })
	testkit.AssertNotNil(t, tp)
}

// ---- Span helpers ----

func TestStart(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("t")
	ctx := context.Background()
	ctx2, span := StartWithTracer(ctx, tr, "test-op")
	defer span.End()
	if ctx2 == ctx {
		t.Error("expected new context with span")
	}
	testkit.RequireNotNil(t, span)
}

func TestRecordError_Nil(t *testing.T) {
	tp := newTestProvider(t)
	_, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	// Should not panic.
	RecordError(span, nil)
}

func TestRecordError_NonNil(t *testing.T) {
	tp := newTestProvider(t)
	_, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	RecordError(span, errTest)
}

var errTest = errorString("test error")

type errorString string

func (e errorString) Error() string { return string(e) }

func TestSetAttributes(t *testing.T) {
	tp := newTestProvider(t)
	_, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	// Should not panic.
	SetAttributes(span, attribute.String("key", "value"), attribute.Int("n", 42))
}

func TestTraceID_NoSpan(t *testing.T) {
	id := TraceID(context.Background())
	testkit.AssertEqual(t, id, "")
}

func TestSpanID_NoSpan(t *testing.T) {
	id := SpanID(context.Background())
	testkit.AssertEqual(t, id, "")
}

func TestTraceID_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	id := TraceID(ctx)
	testkit.AssertNotEqual(t, id, "")
	testkit.AssertLen(t, id, 32)
}

func TestSpanID_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	id := SpanID(ctx)
	testkit.AssertNotEqual(t, id, "")
	testkit.AssertLen(t, id, 16)
}

func TestIsRecording_NoSpan(t *testing.T) {
	testkit.AssertFalse(t, IsRecording(context.Background()))
}

func TestIsRecording_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	testkit.AssertTrue(t, IsRecording(ctx))
}

// ---- HTTP middleware ----

func TestMiddleware_200(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("http")
	handler := Middleware(tr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	handler.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusOK)
}

func TestMiddleware_500(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("http")
	handler := Middleware(tr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/boom", nil)
	handler.ServeHTTP(rr, req)
	testkit.AssertEqual(t, rr.Code, http.StatusInternalServerError)
}

func TestMiddleware_PropagatesContext(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("http")
	var gotTraceID string
	handler := Middleware(tr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotTraceID = TraceID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/trace", nil)
	handler.ServeHTTP(rr, req)
	testkit.AssertNotEqual(t, gotTraceID, "")
}

func TestInjectExtractHTTP(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "parent")
	defer span.End()

	req, _ := http.NewRequest("GET", "http://example.com/", nil)
	InjectHTTP(ctx, req)

	extracted := ExtractHTTP(context.Background(), req)
	_ = extracted // Should not panic; propagation headers set.
}

func TestStart_GlobalTracer(t *testing.T) {
	newTestProvider(t)
	ctx, span := Start(context.Background(), "global-op")
	defer span.End()
	testkit.RequireNotNil(t, span)
	_ = ctx
}

func TestNoopExporter(t *testing.T) {
	n := &noopExporter{}
	testkit.AssertNoError(t, n.ExportSpans(context.Background(), nil))
	testkit.AssertNoError(t, n.Shutdown(context.Background()))
}

func TestStatusWriter_DefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rr, status: http.StatusOK}
	sw.WriteHeader(http.StatusCreated)
	testkit.AssertEqual(t, sw.status, http.StatusCreated)
}

func TestExporterType_Constants(t *testing.T) {
	testkit.AssertNotEqual(t, ExporterStdout, ExporterNoop)
	testkit.AssertContains(t, string(ExporterStdout), "stdout")
}

func TestShutdown_NilProvider(t *testing.T) {
	var p *Provider
	testkit.AssertNoError(t, p.Shutdown(context.Background()))
}

func TestShutdown_NilTP(t *testing.T) {
	p := &Provider{tp: nil}
	testkit.AssertNoError(t, p.Shutdown(context.Background()))
}

func TestShutdown_CancelledContext(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    ExporterNoop,
	})
	testkit.RequireNoError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	// A cancelled context may cause the underlying provider shutdown to fail.
	_ = tp.Shutdown(ctx)
}

func TestStartWithTracer_SpanKind(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("t")
	ctx, span := StartWithTracer(context.Background(), tr, "client-call",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	_ = ctx
	testkit.RequireNotNil(t, span)
}
