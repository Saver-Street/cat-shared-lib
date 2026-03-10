package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	if err != nil {
		t.Fatalf("NewProvider: %v", err)
	}
	t.Cleanup(func() { tp.Shutdown(context.Background()) })
	return tp
}

func TestNewProvider_Noop(t *testing.T) {
	tp := newTestProvider(t)
	if tp == nil {
		t.Fatal("expected non-nil provider")
	}
}

func TestNewProvider_Stdout(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "test",
		Exporter:    ExporterStdout,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tp.Shutdown(context.Background())
}

func TestNewProvider_UnknownExporter(t *testing.T) {
	_, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    "kafka",
	})
	if err == nil {
		t.Fatal("expected error for unknown exporter")
	}
}

func TestNewProvider_WithEnvironment(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Environment: "production",
		Exporter:    ExporterNoop,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tp.Shutdown(context.Background())
}

func TestNewProvider_SampleRateZeroDefaults(t *testing.T) {
	cfg := Config{ServiceName: "s", SampleRate: 0}
	cfg.defaults()
	if cfg.SampleRate != 1.0 {
		t.Errorf("expected SampleRate=1.0, got %f", cfg.SampleRate)
	}
}

func TestNewProvider_ExporterDefault(t *testing.T) {
	cfg := Config{ServiceName: "s"}
	cfg.defaults()
	if cfg.Exporter != ExporterNoop {
		t.Errorf("expected noop exporter, got %q", cfg.Exporter)
	}
}

func TestProvider_Tracer(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("test-component")
	if tr == nil {
		t.Fatal("expected non-nil tracer")
	}
}

func TestProvider_Shutdown(t *testing.T) {
	tp, err := NewProvider(context.Background(), Config{
		ServiceName: "svc",
		Exporter:    ExporterNoop,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tp.Shutdown(context.Background()); err != nil {
		t.Fatalf("unexpected error on shutdown: %v", err)
	}
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
	if span == nil {
		t.Fatal("expected non-nil span")
	}
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
	if id != "" {
		t.Errorf("expected empty trace ID, got %q", id)
	}
}

func TestSpanID_NoSpan(t *testing.T) {
	id := SpanID(context.Background())
	if id != "" {
		t.Errorf("expected empty span ID, got %q", id)
	}
}

func TestTraceID_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	id := TraceID(ctx)
	if id == "" {
		t.Error("expected non-empty trace ID with active span")
	}
	if len(id) != 32 {
		t.Errorf("trace ID should be 32 hex chars, got %d (%q)", len(id), id)
	}
}

func TestSpanID_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	id := SpanID(ctx)
	if id == "" {
		t.Error("expected non-empty span ID with active span")
	}
	if len(id) != 16 {
		t.Errorf("span ID should be 16 hex chars, got %d (%q)", len(id), id)
	}
}

func TestIsRecording_NoSpan(t *testing.T) {
	if IsRecording(context.Background()) {
		t.Error("expected false for context with no span")
	}
}

func TestIsRecording_WithSpan(t *testing.T) {
	tp := newTestProvider(t)
	ctx, span := tp.Tracer("t").Start(context.Background(), "op")
	defer span.End()
	if !IsRecording(ctx) {
		t.Error("expected true for context with active recording span")
	}
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
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
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
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
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
	if gotTraceID == "" {
		t.Error("expected trace ID to be set in handler context")
	}
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
	if span == nil {
		t.Fatal("expected non-nil span")
	}
	_ = ctx
}

func TestNoopExporter(t *testing.T) {
	n := &noopExporter{}
	if err := n.ExportSpans(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.Shutdown(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStatusWriter_DefaultStatus(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusWriter{ResponseWriter: rr, status: http.StatusOK}
	sw.WriteHeader(http.StatusCreated)
	if sw.status != http.StatusCreated {
		t.Errorf("expected 201, got %d", sw.status)
	}
}

func TestExporterType_Constants(t *testing.T) {
	if ExporterStdout == ExporterNoop {
		t.Error("exporter constants must differ")
	}
	if !strings.Contains(string(ExporterStdout), "stdout") {
		t.Error("stdout exporter constant should contain 'stdout'")
	}
}

func TestStartWithTracer_SpanKind(t *testing.T) {
	tp := newTestProvider(t)
	tr := tp.Tracer("t")
	ctx, span := StartWithTracer(context.Background(), tr, "client-call",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()
	_ = ctx
	if span == nil {
		t.Fatal("nil span")
	}
}
