package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestCorrelationID_GeneratesWhenMissing(t *testing.T) {
	var gotID string
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetCorrelationID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertNotEqual(t, gotID, "")
	testkit.AssertLen(t, gotID, 32)
	testkit.AssertEqual(t, rr.Header().Get(CorrelationIDHeader), gotID)
}

func TestCorrelationID_ReusesExisting(t *testing.T) {
	existingID := "existing-correlation-id-456"
	var gotID string
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetCorrelationID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, existingID)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, gotID, existingID)
	testkit.AssertEqual(t, rr.Header().Get(CorrelationIDHeader), existingID)
}

func TestCorrelationID_RejectsOversized(t *testing.T) {
	oversized := strings.Repeat("x", maxRequestIDLen+1)
	var gotID string
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetCorrelationID(r)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, oversized)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertNotEqual(t, gotID, oversized)
	testkit.AssertLen(t, gotID, 32)
	testkit.AssertEqual(t, rr.Header().Get(CorrelationIDHeader), gotID)
}

func TestCorrelationID_ResponseHeader(t *testing.T) {
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	respHeader := rr.Header().Get(CorrelationIDHeader)
	testkit.AssertNotEqual(t, respHeader, "")
	testkit.AssertLen(t, respHeader, 32)
}

func TestCorrelationID_AvailableViaGetCorrelationID(t *testing.T) {
	const incoming = "trace-across-services"
	var gotID string
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetCorrelationID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, incoming)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, gotID, incoming)
}

func TestCorrelationIDFromContext(t *testing.T) {
	ctx := SetCorrelationID(context.Background(), "ctx-corr-789")
	got := CorrelationIDFromContext(ctx)
	testkit.AssertEqual(t, got, "ctx-corr-789")
}

func TestCorrelationIDFromContext_Empty(t *testing.T) {
	got := CorrelationIDFromContext(context.Background())
	testkit.AssertEqual(t, got, "")
}

func TestSetCorrelationID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := SetCorrelationID(req.Context(), "manual-corr-id")
	req = req.WithContext(ctx)

	testkit.AssertEqual(t, GetCorrelationID(req), "manual-corr-id")
}

func TestCorrelationID_EmptyHeaderGeneratesNew(t *testing.T) {
	var gotID string
	handler := CorrelationID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = GetCorrelationID(r)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, "")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertNotEqual(t, gotID, "")
	testkit.AssertLen(t, gotID, 32)
}

func TestCorrelationID_DoesNotInterfereWithRequestID(t *testing.T) {
	var gotCorrID, gotReqID string
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCorrID = GetCorrelationID(r)
		gotReqID = GetRequestID(r)
	})

	handler := CorrelationID(RequestID(inner))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(CorrelationIDHeader, "corr-123")
	req.Header.Set(RequestIDHeader, "req-456")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	testkit.AssertEqual(t, gotCorrID, "corr-123")
	testkit.AssertEqual(t, gotReqID, "req-456")
	testkit.AssertNotEqual(t, gotCorrID, gotReqID)
}

func TestCorrelationID_ContextPropagation(t *testing.T) {
	ctx := context.Background()
	ctx = SetCorrelationID(ctx, "propagated-id")

	got := CorrelationIDFromContext(ctx)
	testkit.AssertEqual(t, got, "propagated-id")

	// Overwrite with a new value.
	ctx = SetCorrelationID(ctx, "updated-id")
	got = CorrelationIDFromContext(ctx)
	testkit.AssertEqual(t, got, "updated-id")
}
