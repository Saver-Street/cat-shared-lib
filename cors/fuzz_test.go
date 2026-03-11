package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func FuzzMiddleware_Origin(f *testing.F) {
	f.Add("https://example.com")
	f.Add("http://localhost:3000")
	f.Add("")
	f.Add("https://EXAMPLE.COM")
	f.Add("null")
	f.Add("https://evil.com")
	f.Add("https://example.com\r\nX-Injected: true")
	f.Fuzz(func(t *testing.T, origin string) {
		handler := Middleware(Config{
			AllowedOrigins: []string{"https://example.com", "http://localhost:3000"},
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		// Must not panic on any origin value.
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})
}

func FuzzMiddleware_Preflight(f *testing.F) {
	f.Add("https://example.com", "POST")
	f.Add("http://localhost", "DELETE")
	f.Add("", "GET")
	f.Add("https://evil.com", "PATCH")
	f.Fuzz(func(t *testing.T, origin, method string) {
		handler := Middleware(Config{
			AllowedOrigins: []string{"*"},
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", origin)
		req.Header.Set("Access-Control-Request-Method", method)
		rec := httptest.NewRecorder()

		// Must not panic on any combination.
		handler.ServeHTTP(rec, req)
	})
}

func FuzzMiddleware_WildcardCredentials(f *testing.F) {
	f.Add("https://app.example.com")
	f.Add("")
	f.Add("null")
	f.Fuzz(func(t *testing.T, origin string) {
		handler := Middleware(Config{
			AllowedOrigins:   []string{"*"},
			AllowCredentials: true,
		})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// With credentials + wildcard, origin should be echoed back, not "*".
		if origin != "" {
			acao := rec.Header().Get("Access-Control-Allow-Origin")
			if acao == "*" {
				t.Error("should not use * with credentials")
			}
		}
	})
}
