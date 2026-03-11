package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestContentType_Allowed(t *testing.T) {
	handler := ContentType("application/json")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestContentType_Rejected(t *testing.T) {
	handler := ContentType("application/json")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusUnsupportedMediaType)
}

func TestContentType_GetPassthrough(t *testing.T) {
	handler := ContentType("application/json")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusOK)
}

func TestContentType_MultipleAllowed(t *testing.T) {
	handler := ContentType("application/json", "application/xml")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPut, "/", nil)
	req.Header.Set("Content-Type", "application/xml")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	testkit.AssertStatus(t, rr, http.StatusOK)
}
