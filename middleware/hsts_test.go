package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHSTS(t *testing.T) {
	handler := HSTS(31536000, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
}

func TestHSTS_IncludeSubDomains(t *testing.T) {
	handler := HSTS(63072000, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Header().Get("Strict-Transport-Security"), "max-age=63072000; includeSubDomains")
}

func BenchmarkHSTS(b *testing.B) {
	handler := HSTS(31536000, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}
