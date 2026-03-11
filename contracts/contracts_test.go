package contracts

import (
	"context"
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
		testkit.AssertEqual(t, h.IsHealthy(), tt.want)
	}
}

func TestHealthState_Constants(t *testing.T) {
	testkit.AssertEqual(t, string(HealthStateOK), "ok")
	testkit.AssertEqual(t, string(HealthStateDegraded), "degraded")
	testkit.AssertEqual(t, string(HealthStateDown), "down")
}

// ---- StandardError tests ----

func TestNewStandardError(t *testing.T) {
	e := NewStandardError("NOT_FOUND", "resource not found")
	testkit.AssertEqual(t, e.Code, "NOT_FOUND")
	testkit.AssertEqual(t, e.Message, "resource not found")
	testkit.AssertNil(t, e.Details)
}

func TestNewStandardErrorWithDetails(t *testing.T) {
	details := map[string]any{"field": "email", "reason": "invalid format"}
	e := NewStandardErrorWithDetails("VALIDATION_ERROR", "validation failed", details)
	testkit.AssertEqual(t, e.Code, "VALIDATION_ERROR")
	if e.Details == nil {
		t.Fatal("expected Details to be set")
	}
	testkit.AssertEqual(t, e.Details["field"], "email")
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
			testkit.AssertEqual(t, tt.e.Error(), tt.wantMsg)
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
	testkit.AssertTrue(t, status.IsHealthy())
	testkit.AssertEqual(t, status.Service, "test-svc")
}

func TestMockService_Info(t *testing.T) {
	svc := &mockService{name: "jobs-service", version: "2.0.0", env: "production"}
	testkit.AssertEqual(t, svc.Name(), "jobs-service")
	testkit.AssertEqual(t, svc.Version(), "2.0.0")
	testkit.AssertEqual(t, svc.Environment(), "production")
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
	testkit.AssertFalse(t, h.IsHealthy())
	testkit.AssertEqual(t, h.Checks["database"], "connection refused")
}
