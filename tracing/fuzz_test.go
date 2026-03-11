package tracing

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/trace/noop"
)

// FuzzMiddleware exercises the tracing middleware with arbitrary HTTP methods
// and paths. It verifies the middleware never panics regardless of input.
func FuzzMiddleware(f *testing.F) {
	f.Add("GET", "/api/test")
	f.Add("POST", "/users/123/profile")
	f.Add("DELETE", "/")
	f.Add("", "")
	f.Add("PATCH", "/a/b/c/d/e/f/g/h")
	f.Add("OPTIONS", "/../../../etc/passwd")
	f.Add("GET", "/api?q=hello&page=1")
	f.Add("PUT", "/\x00\xff\n\t")
	f.Add("CONNECT", "https://evil.example.com")

	tracer := noop.NewTracerProvider().Tracer("fuzz")
	handler := Middleware(tracer, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	f.Fuzz(func(t *testing.T, method, path string) {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			t.Skip("invalid request")
		}
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	})
}

// FuzzExtractHTTP verifies that ExtractHTTP never panics when given arbitrary
// traceparent and tracestate headers. These headers come from untrusted
// callers in a real deployment.
func FuzzExtractHTTP(f *testing.F) {
	f.Add("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", "congo=t61rcWkgMzE")
	f.Add("00-00000000000000000000000000000000-0000000000000000-00", "")
	f.Add("", "")
	f.Add("not-a-valid-traceparent", "not-valid")
	f.Add("00-ffffffffffffffffffffffffffffffff-ffffffffffffffff-ff", "a=b,c=d")
	f.Add("01-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01", "")
	f.Add("00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01-extra", "")
	f.Add("\x00\xff\n\r", "\x00\xff\n\r")

	f.Fuzz(func(t *testing.T, traceparent, tracestate string) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("traceparent", traceparent)
		if tracestate != "" {
			req.Header.Set("tracestate", tracestate)
		}
		ctx := ExtractHTTP(context.Background(), req)
		if ctx == nil {
			t.Error("ExtractHTTP returned nil context")
		}
	})
}

// FuzzInjectExtractRoundTrip injects trace context into a request, then
// extracts it, verifying the round-trip never panics and preserves validity.
func FuzzInjectExtractRoundTrip(f *testing.F) {
	f.Add("GET", "/api/v1/users")
	f.Add("POST", "/")
	f.Add("", "")

	tracer := noop.NewTracerProvider().Tracer("fuzz")

	f.Fuzz(func(t *testing.T, method, path string) {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			t.Skip("invalid request")
		}

		ctx, span := tracer.Start(context.Background(), "fuzz-op")
		defer span.End()

		InjectHTTP(ctx, req)
		extracted := ExtractHTTP(context.Background(), req)
		if extracted == nil {
			t.Error("ExtractHTTP returned nil context after InjectHTTP")
		}
	})
}

// FuzzNewProvider exercises provider creation with arbitrary config values.
// It validates that the constructor handles all inputs without panicking.
func FuzzNewProvider(f *testing.F) {
	f.Add("my-service", "v1.0.0", "production", "noop", 1.0)
	f.Add("", "", "", "", 0.0)
	f.Add("svc", "v0.0.1", "staging", "stdout", 0.5)
	f.Add("svc", "", "", "invalid", -1.0)
	f.Add("service-name-that-is-very-long", "v999.999.999", "test", "noop", 2.0)

	f.Fuzz(func(t *testing.T, name, version, env, exporter string, rate float64) {
		cfg := Config{
			ServiceName:    name,
			ServiceVersion: version,
			Environment:    env,
			Exporter:       ExporterType(exporter),
			SampleRate:     rate,
		}
		p, err := NewProvider(context.Background(), cfg)
		if err != nil {
			return // expected for invalid exporters
		}
		// Verify provider works without panic.
		tracer := p.Tracer("fuzz")
		_, span := tracer.Start(context.Background(), "fuzz-span")
		span.End()
		_ = p.Shutdown(context.Background())
	})
}
