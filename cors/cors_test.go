package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestDefaults_WildcardOrigin(t *testing.T) {
	handler := Middleware(Config{})(newHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected wildcard origin, got %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestNoOriginHeader_Passthrough(t *testing.T) {
	handler := Middleware(Config{})(newHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS headers without Origin, got %q", got)
	}
}

func TestSpecificOrigins_Allowed(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com", "http://admin.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://app.example.com" {
		t.Errorf("expected specific origin echo, got %q", got)
	}
	if got := rr.Header().Get("Vary"); got != "Origin" {
		t.Errorf("expected Vary: Origin, got %q", got)
	}
}

func TestSpecificOrigins_Denied(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS header for disallowed origin, got %q", got)
	}
}

func TestOriginMatchIsCaseInsensitive(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://APP.Example.COM"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://app.example.com" {
		t.Errorf("expected origin match, got %q", got)
	}
}

func TestPreflight_OptionsRequest(t *testing.T) {
	handler := Middleware(Config{})(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204 for preflight, got %d", rr.Code)
	}
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
	if got := rr.Header().Get("Access-Control-Max-Age"); got != "86400" {
		t.Errorf("expected max-age 86400, got %q", got)
	}
}

func TestCredentials(t *testing.T) {
	cfg := Config{
		AllowedOrigins:   []string{"http://app.example.com"},
		AllowCredentials: true,
	}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected credentials header, got %q", got)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Error("should not use wildcard origin with credentials")
	}
}

func TestWildcardWithCredentials_EchosOrigin(t *testing.T) {
	cfg := Config{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://any.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://any.example.com" {
		t.Errorf("expected echoed origin, got %q", got)
	}
}

func TestExposedHeaders(t *testing.T) {
	cfg := Config{ExposedHeaders: []string{"X-Total-Count", "X-Request-ID"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Expose-Headers"); got != "X-Total-Count, X-Request-ID" {
		t.Errorf("expected exposed headers, got %q", got)
	}
}

func TestCustomMethods(t *testing.T) {
	cfg := Config{AllowedMethods: []string{http.MethodGet, http.MethodPost}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST" {
		t.Errorf("expected custom methods, got %q", got)
	}
}

func TestCustomHeaders(t *testing.T) {
	cfg := Config{AllowedHeaders: []string{"X-Custom"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Headers"); got != "X-Custom" {
		t.Errorf("expected custom headers, got %q", got)
	}
}

func TestCustomMaxAge(t *testing.T) {
	cfg := Config{MaxAge: 3600}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Max-Age"); got != "3600" {
		t.Errorf("expected 3600, got %q", got)
	}
}

func TestPreflightDisallowed_NoHeaders(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Methods"); got != "" {
		t.Errorf("expected no CORS headers for disallowed origin preflight, got methods %q", got)
	}
}

func BenchmarkMiddleware(b *testing.B) {
	handler := Middleware(Config{})(newHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
