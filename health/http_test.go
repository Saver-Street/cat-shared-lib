package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHTTPChecker_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "cat-shared-lib/health-checker" {
			t.Errorf("User-Agent = %q, want cat-shared-lib/health-checker", r.Header.Get("User-Agent"))
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "remote-api",
		URL:  srv.URL,
	})

	if checker.Name() != "remote-api" {
		t.Errorf("Name() = %q, want remote-api", checker.Name())
	}

	err := checker.Check(t.Context())
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
}

func TestHTTPChecker_Unhealthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"degraded"}`))
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "remote-api",
		URL:  srv.URL,
	})

	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for 503")
	}
}

func TestHTTPChecker_CustomStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	checker := HTTPChecker(HTTPCheckerConfig{
		Name:           "custom",
		URL:            srv.URL,
		ExpectedStatus: http.StatusNoContent,
	})

	err := checker.Check(t.Context())
	if err != nil {
		t.Errorf("Check() error = %v, want nil for 204 with custom expected", err)
	}
}

func TestHTTPChecker_ConnectionRefused(t *testing.T) {
	checker := HTTPChecker(HTTPCheckerConfig{
		Name: "down-service",
		URL:  "http://127.0.0.1:1", // unlikely to be listening
	})

	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for connection refused")
	}
}

func TestAggregateChecker_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") != "cat-shared-lib/health-aggregator" {
			t.Errorf("User-Agent = %q, want cat-shared-lib/health-aggregator", r.Header.Get("User-Agent"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "ok", Service: "downstream"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	if err != nil {
		t.Errorf("Check() error = %v", err)
	}
}

func TestAggregateChecker_Degraded(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(Status{Status: "degraded"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for degraded downstream")
	}
}

func TestAggregateChecker_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// 200 but status field is not "ok"
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "degraded"})
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for non-ok status field")
	}
}

func TestAggregateChecker_InvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`not valid json`))
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for invalid JSON response")
	}
}

func TestAggregateChecker_ConnectionRefused(t *testing.T) {
	checker := AggregateChecker("down", "http://127.0.0.1:1")
	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for connection refused")
	}
}

func TestAggregateChecker_UnexpectedStatusCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	checker := AggregateChecker("downstream", srv.URL)
	err := checker.Check(t.Context())
	if err == nil {
		t.Error("Check() = nil, want error for 500 status")
	}
	if !strings.Contains(err.Error(), "returned 500") {
		t.Errorf("error = %v, want mention of 500", err)
	}
}

func TestHTTPChecker_InHandler(t *testing.T) {
	// Test that HTTPChecker integrates with the existing Handler.
	remoteSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer remoteSrv.Close()

	handler := Handler("gateway", "1.0.0",
		HTTPChecker(HTTPCheckerConfig{Name: "billing", URL: remoteSrv.URL}),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var status Status
	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Status != "ok" {
		t.Errorf("status = %q, want ok", status.Status)
	}
	if status.Checks["billing"] != "ok" {
		t.Errorf("checks[billing] = %q, want ok", status.Checks["billing"])
	}
}

func TestHTTPChecker_DegradedInHandler(t *testing.T) {
	remoteSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer remoteSrv.Close()

	handler := Handler("gateway", "1.0.0",
		HTTPChecker(HTTPCheckerConfig{Name: "billing", URL: remoteSrv.URL}),
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}

	var status Status
	_ = json.NewDecoder(rec.Body).Decode(&status)
	if status.Status != "degraded" {
		t.Errorf("status = %q, want degraded", status.Status)
	}
}
