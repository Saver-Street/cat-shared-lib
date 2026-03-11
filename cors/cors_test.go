package cors

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "*")
	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestNoOriginHeader_Passthrough(t *testing.T) {
	handler := Middleware(Config{})(newHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "")
}

func TestSpecificOrigins_Allowed(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com", "http://admin.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "http://app.example.com")
	testkit.AssertHeader(t, rr, "Vary", "Origin")
}

func TestSpecificOrigins_Denied(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "")
}

func TestOriginMatchIsCaseInsensitive(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://APP.Example.COM"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://app.example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "http://app.example.com")
}

func TestPreflight_OptionsRequest(t *testing.T) {
	handler := Middleware(Config{})(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertStatus(t, rr, http.StatusNoContent)
	if got := rr.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
	if got := rr.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
	testkit.AssertHeader(t, rr, "Access-Control-Max-Age", "86400")
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

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Credentials", "true")
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

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Origin", "http://any.example.com")
}

func TestExposedHeaders(t *testing.T) {
	cfg := Config{ExposedHeaders: []string{"X-Total-Count", "X-Request-ID"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Expose-Headers", "X-Total-Count, X-Request-ID")
}

func TestCustomMethods(t *testing.T) {
	cfg := Config{AllowedMethods: []string{http.MethodGet, http.MethodPost}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Methods", "GET, POST")
}

func TestCustomHeaders(t *testing.T) {
	cfg := Config{AllowedHeaders: []string{"X-Custom"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Headers", "X-Custom")
}

func TestCustomMaxAge(t *testing.T) {
	cfg := Config{MaxAge: 3600}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://example.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Max-Age", "3600")
}

func TestPreflightDisallowed_NoHeaders(t *testing.T) {
	cfg := Config{AllowedOrigins: []string{"http://app.example.com"}}
	handler := Middleware(cfg)(newHandler())

	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertHeader(t, rr, "Access-Control-Allow-Methods", "")
}

func BenchmarkMiddleware(b *testing.B) {
	handler := Middleware(Config{})(newHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://example.com")

	b.ResetTimer()
	for b.Loop() {
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}
