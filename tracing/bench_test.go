package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"
)

func BenchmarkMiddleware(b *testing.B) {
	tracer := noop.NewTracerProvider().Tracer("bench")
	handler := Middleware(tracer, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func BenchmarkStart(b *testing.B) {
	ctx := context.Background()
	for b.Loop() {
		_, span := Start(ctx, "bench-op")
		span.End()
	}
}

func BenchmarkTraceID(b *testing.B) {
	ctx := context.Background()
	for b.Loop() {
		TraceID(ctx)
	}
}

func BenchmarkRecordError_Nil(b *testing.B) {
	tracer := noop.NewTracerProvider().Tracer("bench")
	_, span := tracer.Start(context.Background(), "op")
	defer span.End()
	b.ResetTimer()
	for b.Loop() {
		RecordError(span, nil)
	}
}

func BenchmarkInjectHTTP(b *testing.B) {
	ctx := context.Background()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ResetTimer()
	for b.Loop() {
		InjectHTTP(ctx, req)
	}
}
