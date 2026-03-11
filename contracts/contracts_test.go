package contracts

import (
	"context"
	"net/http"
	"testing"
)

// ---- HealthStatus tests ----

func TestHealthStatus_IsHealthy(t *testing.T) {
	tests := []struct {
		state HealthState
		want  bool
	}{
		{HealthStateOK, true},
		{HealthStateDegraded, false},
		{HealthStateDown, false},
	}
	for _, tt := range tests {
		h := HealthStatus{State: tt.state}
		if got := h.IsHealthy(); got != tt.want {
			t.Errorf("state=%q: IsHealthy()=%v, want %v", tt.state, got, tt.want)
		}
	}
}

func TestHealthState_Constants(t *testing.T) {
	if HealthStateOK != "ok" {
		t.Errorf("HealthStateOK = %q, want %q", HealthStateOK, "ok")
	}
	if HealthStateDegraded != "degraded" {
		t.Errorf("HealthStateDegraded = %q, want %q", HealthStateDegraded, "degraded")
	}
	if HealthStateDown != "down" {
		t.Errorf("HealthStateDown = %q, want %q", HealthStateDown, "down")
	}
}

// ---- StandardError tests ----

func TestNewStandardError(t *testing.T) {
	e := NewStandardError("NOT_FOUND", "resource not found")
	if e.Code != "NOT_FOUND" {
		t.Errorf("Code = %q, want NOT_FOUND", e.Code)
	}
	if e.Message != "resource not found" {
		t.Errorf("Message = %q, want 'resource not found'", e.Message)
	}
	if e.Details != nil {
		t.Errorf("Details should be nil, got %v", e.Details)
	}
}

func TestNewStandardErrorWithDetails(t *testing.T) {
	details := map[string]any{"field": "email", "reason": "invalid format"}
	e := NewStandardErrorWithDetails("VALIDATION_ERROR", "validation failed", details)
	if e.Code != "VALIDATION_ERROR" {
		t.Errorf("Code = %q, want VALIDATION_ERROR", e.Code)
	}
	if e.Details == nil {
		t.Fatal("expected Details to be set")
	}
	if e.Details["field"] != "email" {
		t.Errorf("Details[field] = %v, want email", e.Details["field"])
	}
}

func TestStandardError_Error(t *testing.T) {
	tests := []struct {
		name    string
		e       *StandardError
		wantMsg string
	}{
		{
			name:    "with message",
			e:       NewStandardError("NOT_FOUND", "user not found"),
			wantMsg: "NOT_FOUND: user not found",
		},
		{
			name:    "code only",
			e:       &StandardError{Code: "INTERNAL_ERROR"},
			wantMsg: "INTERNAL_ERROR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.wantMsg {
				t.Errorf("Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

// ---- Interface compliance tests ----

// mockService implements the full Service interface for compile-time validation.
type mockService struct {
	name    string
	version string
	env     string
}

func (m *mockService) HealthCheck(_ context.Context) (HealthStatus, error) {
	return HealthStatus{State: HealthStateOK, Service: m.name, Version: m.version}, nil
}

func (m *mockService) Name() string                    { return m.name }
func (m *mockService) Version() string                 { return m.version }
func (m *mockService) Environment() string             { return m.env }
func (m *mockService) RegisterRoutes(_ *http.ServeMux) {}

// Compile-time assertions: mockService satisfies all contracts.
var (
	_ ServiceHealth = (*mockService)(nil)
	_ ServiceInfo   = (*mockService)(nil)
	_ Handler       = (*mockService)(nil)
	_ Service       = (*mockService)(nil)
)

func TestMockService_HealthCheck(t *testing.T) {
	svc := &mockService{name: "test-svc", version: "1.0.0", env: "test"}
	status, err := svc.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.IsHealthy() {
		t.Errorf("expected healthy status, got %q", status.State)
	}
	if status.Service != "test-svc" {
		t.Errorf("Service = %q, want test-svc", status.Service)
	}
}

func TestMockService_Info(t *testing.T) {
	svc := &mockService{name: "jobs-service", version: "2.0.0", env: "production"}
	if got := svc.Name(); got != "jobs-service" {
		t.Errorf("Name() = %q, want jobs-service", got)
	}
	if got := svc.Version(); got != "2.0.0" {
		t.Errorf("Version() = %q, want 2.0.0", got)
	}
	if got := svc.Environment(); got != "production" {
		t.Errorf("Environment() = %q, want production", got)
	}
}

func TestMockService_RegisterRoutes(t *testing.T) {
	svc := &mockService{}
	mux := http.NewServeMux()
	// Should not panic.
	svc.RegisterRoutes(mux)
}

func TestHealthStatus_Checks(t *testing.T) {
	h := HealthStatus{
		State:   HealthStateDegraded,
		Service: "api",
		Version: "1.0.0",
		Checks: map[string]string{
			"database": "connection refused",
			"cache":    "ok",
		},
	}
	if h.IsHealthy() {
		t.Error("expected degraded to not be healthy")
	}
	if h.Checks["database"] != "connection refused" {
		t.Errorf("unexpected database check: %q", h.Checks["database"])
	}
}
