package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestRequestID_GeneratesNew(t *testing.T) {
	var gotID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetRequestID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertNotEqual(t, gotID, "")
	testkit.AssertLen(t, gotID, 32)
	testkit.AssertEqual(t, rr.Header().Get(RequestIDHeader), gotID)
}

func TestRequestID_ReusesExisting(t *testing.T) {
	existingID := "existing-request-id-123"
	var gotID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetRequestID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(RequestIDHeader, existingID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, gotID, existingID)
	testkit.AssertEqual(t, rr.Header().Get(RequestIDHeader), existingID)
}

func TestRequestID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := GetRequestID(r)
		if ids[id] {
			t.Errorf("duplicate ID: %s", id)
		}
		ids[id] = true
	}))

	for range 100 {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}
}

func TestSetRequestID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetRequestID(req.Context(), "manual-id")
	req = req.WithContext(ctx)

	testkit.AssertEqual(t, GetRequestID(req), "manual-id")
}

func TestGetRequestID_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	testkit.AssertEqual(t, GetRequestID(req), "")
}
