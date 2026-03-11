package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rr, req)

	if captured == "" {
		t.Fatal("expected generated request ID in context")
	}
	if len(captured) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32-char hex ID, got %d chars: %s", len(captured), captured)
	}
	if rr.Header().Get(RequestIDHeader) != captured {
		t.Errorf("response header = %s, want %s", rr.Header().Get(RequestIDHeader), captured)
	}
}

func TestRequestID_ReusesIncoming(t *testing.T) {
	const incoming = "my-trace-id-123"
	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, incoming)
	handler.ServeHTTP(rr, req)

	if captured != incoming {
		t.Errorf("got %s, want %s", captured, incoming)
	}
	if rr.Header().Get(RequestIDHeader) != incoming {
		t.Errorf("response header = %s, want %s", rr.Header().Get(RequestIDHeader), incoming)
	}
}

func TestRequestID_RejectsOversized(t *testing.T) {
	oversized := strings.Repeat("x", maxRequestIDLen+1)
	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, oversized)
	handler.ServeHTTP(rr, req)

	if captured == oversized {
		t.Error("expected oversized request ID to be replaced")
	}
	if len(captured) != 32 {
		t.Errorf("expected 32-char generated ID, got %d", len(captured))
	}
}

func TestRequestID_AcceptsMaxLength(t *testing.T) {
	maxLen := strings.Repeat("a", maxRequestIDLen)
	var captured string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = GetRequestID(r)
	}))

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, maxLen)
	handler.ServeHTTP(rr, req)

	if captured != maxLen {
		t.Errorf("expected max-length ID to be accepted")
	}
	_ = rr
}

func TestSetRequestID_GetRequestID(t *testing.T) {
	ctx := SetRequestID(context.Background(), "test-id-456")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(ctx)

	got := GetRequestID(req)
	if got != "test-id-456" {
		t.Errorf("got %s, want test-id-456", got)
	}
}

func TestRequestIDFromContext(t *testing.T) {
	ctx := SetRequestID(context.Background(), "ctx-id-789")
	got := RequestIDFromContext(ctx)
	if got != "ctx-id-789" {
		t.Errorf("got %s, want ctx-id-789", got)
	}
}

func TestRequestIDFromContext_Empty(t *testing.T) {
	got := RequestIDFromContext(context.Background())
	if got != "" {
		t.Errorf("expected empty string, got %s", got)
	}
}

func TestGenerateID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool, 100)
	for i := range 100 {
		id := generateID()
		if seen[id] {
			t.Fatalf("duplicate ID at iteration %d: %s", i, id)
		}
		seen[id] = true
	}
}
