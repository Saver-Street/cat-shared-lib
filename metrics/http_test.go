package metrics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestHTTPMiddleware_CountsRequests(t *testing.T) {
	reg := NewRegistry()
	mw := HTTPMiddleware(reg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))

	for range 3 {
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
		handler.ServeHTTP(rr, req)
		testkit.AssertStatus(t, rr, http.StatusOK)
	}

	output := reg.Expose()
	testkit.AssertContains(t, output, "http_requests_total")
	testkit.AssertContains(t, output, `method="GET"`)
	testkit.AssertContains(t, output, `status="200"`)
}

func TestHTTPMiddleware_RecordsDuration(t *testing.T) {
	reg := NewRegistry()
	mw := HTTPMiddleware(reg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/items", nil)
	handler.ServeHTTP(rr, req)

	output := reg.Expose()
	testkit.AssertContains(t, output, "http_request_duration_seconds")
	testkit.AssertContains(t, output, "_count 1")
}

func TestHTTPMiddleware_TracksDifferentStatuses(t *testing.T) {
	reg := NewRegistry()
	mw := HTTPMiddleware(reg)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ok" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/ok", nil))
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/missing", nil))

	output := reg.Expose()
	testkit.AssertContains(t, output, `status="200"`)
	testkit.AssertContains(t, output, `status="404"`)
}

func TestStatusCapturer_DefaultCode(t *testing.T) {
	rr := httptest.NewRecorder()
	sw := &statusCapturer{ResponseWriter: rr, code: http.StatusOK}
	// Write without calling WriteHeader
	sw.Write([]byte("hello"))
	testkit.AssertEqual(t, sw.code, http.StatusOK)
}
