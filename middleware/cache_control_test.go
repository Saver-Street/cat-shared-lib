package middleware

import (
"net/http"
"net/http/httptest"
"testing"
"time"

"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestCacheControl_Public(t *testing.T) {
handler := CacheControl(1*time.Hour, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

testkit.AssertEqual(t, w.Header().Get("Cache-Control"), "public, max-age=3600")
}

func TestCacheControl_Private(t *testing.T) {
handler := CacheControl(5*time.Minute, false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

testkit.AssertEqual(t, w.Header().Get("Cache-Control"), "private, max-age=300")
}

func TestCacheControl_Zero(t *testing.T) {
handler := CacheControl(0, true)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

testkit.AssertEqual(t, w.Header().Get("Cache-Control"), "public, max-age=0")
}
