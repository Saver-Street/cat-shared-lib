package middleware

import (
"net/http"
"net/http/httptest"
"testing"

"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestAllowMethods_Allowed(t *testing.T) {
handler := AllowMethods(http.MethodGet, http.MethodPost)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
testkit.AssertEqual(t, w.Code, http.StatusOK)
}

func TestAllowMethods_Rejected(t *testing.T) {
handler := AllowMethods(http.MethodGet)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusOK)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodDelete, "/", nil))
testkit.AssertEqual(t, w.Code, http.StatusMethodNotAllowed)
testkit.AssertEqual(t, w.Header().Get("Allow"), "GET")
}

func TestAllowMethods_POST(t *testing.T) {
handler := AllowMethods(http.MethodPost)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.WriteHeader(http.StatusCreated)
}))

w := httptest.NewRecorder()
handler.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/", nil))
testkit.AssertEqual(t, w.Code, http.StatusCreated)
}
