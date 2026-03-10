package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
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

	if gotID == "" {
		t.Error("expected generated request ID, got empty")
	}
	if len(gotID) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32 hex chars, got %d: %q", len(gotID), gotID)
	}
	if rr.Header().Get(RequestIDHeader) != gotID {
		t.Errorf("response header %q != context value %q", rr.Header().Get(RequestIDHeader), gotID)
	}
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

	if gotID != existingID {
		t.Errorf("got %q, want %q", gotID, existingID)
	}
	if rr.Header().Get(RequestIDHeader) != existingID {
		t.Errorf("response header %q != %q", rr.Header().Get(RequestIDHeader), existingID)
	}
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

	if got := GetRequestID(req); got != "manual-id" {
		t.Errorf("got %q, want %q", got, "manual-id")
	}
}

func TestGetRequestID_Empty(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	if got := GetRequestID(req); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}
