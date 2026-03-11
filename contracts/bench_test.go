package contracts

import (
	"testing"
)

func BenchmarkNewStandardError(b *testing.B) {
	for b.Loop() {
		_ = NewStandardError("NOT_FOUND", "resource not found")
	}
}

func BenchmarkNewStandardErrorWithDetails(b *testing.B) {
	details := map[string]any{"field": "email", "reason": "invalid format"}
	for b.Loop() {
		_ = NewStandardErrorWithDetails("VALIDATION_ERROR", "invalid input", details)
	}
}

func BenchmarkStandardError_Error(b *testing.B) {
	e := NewStandardError("BAD_REQUEST", "invalid input")
	for b.Loop() {
		_ = e.Error()
	}
}

func BenchmarkHealthStatus_IsHealthy(b *testing.B) {
	h := HealthStatus{State: HealthStateOK}
	for b.Loop() {
		_ = h.IsHealthy()
	}
}
