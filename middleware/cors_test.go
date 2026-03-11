package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func corsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestCORS_PreflightAllOrigins(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"*"}})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusNoContent)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "*")
	testkit.AssertContains(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET")
	testkit.AssertContains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}

func TestCORS_PreflightSpecificOrigin(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"https://app.test"}})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.test")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusNoContent)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "https://app.test")
	testkit.AssertEqual(t, rec.Header().Get("Vary"), "Origin")
}

func TestCORS_DisallowedOrigin(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"https://allowed.com"}})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "")
}

func TestCORS_NoOriginHeader(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"*"}})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "")
}

func TestCORS_WithCredentials(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins:   []string{"https://app.test"},
		AllowCredentials: true,
	})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.test")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Credentials"), "true")
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "https://app.test")
}

func TestCORS_CredentialsWithWildcard(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowCredentials: true,
	})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://any.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "https://any.com")
}

func TestCORS_ExposedHeaders(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		ExposedHeaders: []string{"X-Request-ID", "X-Total-Count"},
	})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertContains(t, rec.Header().Get("Access-Control-Expose-Headers"), "X-Request-ID")
}

func TestCORS_CustomMethods(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST"},
	})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Methods"), "GET, POST")
}

func TestCORS_CustomMaxAge(t *testing.T) {
	handler := CORS(CORSConfig{
		AllowedOrigins: []string{"*"},
		MaxAge:         24 * time.Hour,
	})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Max-Age"), "86400")
}

func TestCORS_SimpleRequest(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"*"}})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertEqual(t, rec.Code, http.StatusOK)
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Allow-Origin"), "*")
}

func TestCORS_Defaults(t *testing.T) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"*"}})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	testkit.AssertContains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	testkit.AssertEqual(t, rec.Header().Get("Access-Control-Max-Age"), "300")
}

func BenchmarkCORS_Preflight(b *testing.B) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"*"}})(corsHandler())
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	for b.Loop() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}

func BenchmarkCORS_SimpleRequest(b *testing.B) {
	handler := CORS(CORSConfig{AllowedOrigins: []string{"https://app.test"}})(corsHandler())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.test")
	for b.Loop() {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
