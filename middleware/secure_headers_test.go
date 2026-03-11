package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestSecureHeaders(t *testing.T) {
	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("X-Content-Type-Options"), "nosniff")
	testkit.AssertEqual(t, w.Header().Get("X-Frame-Options"), "DENY")
	testkit.AssertEqual(t, w.Header().Get("Referrer-Policy"), "strict-origin-when-cross-origin")
}

func TestSecureHeaders_PreservesExisting(t *testing.T) {
	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Header().Get("X-Custom"), "value")
	testkit.AssertEqual(t, w.Header().Get("X-Content-Type-Options"), "nosniff")
}

func BenchmarkSecureHeaders(b *testing.B) {
	handler := SecureHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}
