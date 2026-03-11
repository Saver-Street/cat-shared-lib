package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMiddleware_Preflight(b *testing.B) {
	handler := Middleware(Config{
		AllowedOrigins: []string{"https://app.example.com"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodOptions, "/api/data", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_NoOrigin(b *testing.B) {
	handler := Middleware(Config{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkMiddleware_Disallowed(b *testing.B) {
	handler := Middleware(Config{
		AllowedOrigins: []string{"https://trusted.com"},
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/api/data", nil)
	req.Header.Set("Origin", "https://evil.com")

	b.ResetTimer()
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}
