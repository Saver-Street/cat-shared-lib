package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNoCache(t *testing.T) {
	handler := NoCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/api/data", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Cache-Control"), "no-store, no-cache, must-revalidate")
	testkit.AssertEqual(t, w.Header().Get("Pragma"), "no-cache")
	testkit.AssertEqual(t, w.Header().Get("Expires"), "0")
}

func TestNoCache_PreservesExisting(t *testing.T) {
	handler := NoCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "value")
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, r)

	testkit.AssertEqual(t, w.Header().Get("X-Custom"), "value")
	testkit.AssertEqual(t, w.Header().Get("Cache-Control"), "no-store, no-cache, must-revalidate")
}

func BenchmarkNoCache(b *testing.B) {
	handler := NoCache(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	r := httptest.NewRequest("GET", "/", nil)
	for b.Loop() {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
	}
}
