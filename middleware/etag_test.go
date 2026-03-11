package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestETag_SetsHeader(t *testing.T) {
	handler := ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertNotEqual(t, w.Header().Get("ETag"), "")
	testkit.AssertContains(t, w.Body.String(), `"ok"`)
}

func TestETag_NotModified(t *testing.T) {
	handler := ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))

	// First request to get the ETag
	w1 := httptest.NewRecorder()
	handler.ServeHTTP(w1, httptest.NewRequest(http.MethodGet, "/", nil))
	etag := w1.Header().Get("ETag")
	testkit.AssertNotEqual(t, etag, "")

	// Second request with If-None-Match
	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.Header.Set("If-None-Match", etag)
	w2 := httptest.NewRecorder()
	handler.ServeHTTP(w2, r2)

	testkit.AssertEqual(t, w2.Code, http.StatusNotModified)
	testkit.AssertEqual(t, w2.Body.Len(), 0)
}

func TestETag_SkipsNonGET(t *testing.T) {
	handler := ETag(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))

	testkit.AssertEqual(t, w.Code, http.StatusCreated)
	testkit.AssertEqual(t, w.Header().Get("ETag"), "")
}
