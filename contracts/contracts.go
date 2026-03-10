// Package contracts defines the Go interfaces and shared types that all
// microservices in the platform must implement. Importing this package ensures
// compile-time enforcement of the service contracts.
package contracts

import (
	"context"
	"net/http"
)

// HealthState represents the operational state of a service or one of its
// dependencies.
type HealthState string

const (
	// HealthStateOK means the service or dependency is fully operational.
	HealthStateOK HealthState = "ok"
	// HealthStateDegraded means the service is running but one or more
	// dependencies are unhealthy.
	HealthStateDegraded HealthState = "degraded"
	// HealthStateDown means the service cannot serve traffic.
	HealthStateDown HealthState = "down"
)

// HealthStatus is the result returned by ServiceHealth.HealthCheck.
type HealthStatus struct {
	// State is the overall health of the service.
	State HealthState `json:"state"`
	// Service is the canonical service name.
	Service string `json:"service"`
	// Version is the running binary version.
	Version string `json:"version"`
	// Checks contains per-dependency health results keyed by dependency name.
	Checks map[string]string `json:"checks,omitempty"`
}

// IsHealthy reports whether the status represents a fully healthy service.
func (h HealthStatus) IsHealthy() bool { return h.State == HealthStateOK }

// ServiceHealth is the interface that all microservices must implement to
// expose their health-check endpoint.
type ServiceHealth interface {
	// HealthCheck returns the current health of the service and its
	// dependencies. Implementations must honour ctx cancellation / deadlines.
	HealthCheck(ctx context.Context) (HealthStatus, error)
}

// ServiceInfo is the interface that all microservices must implement to
// expose their identity and configuration metadata.
type ServiceInfo interface {
	// Name returns the canonical, URL-safe service name (e.g. "jobs-service").
	Name() string
	// Version returns the running binary version string (e.g. "1.2.3").
	Version() string
	// Environment returns the deployment environment
	// (e.g. "production", "staging", "development").
	Environment() string
}

// Handler is the interface that all microservices must implement to register
// their HTTP routes.
type Handler interface {
	// RegisterRoutes mounts the service's routes on the given ServeMux.
	RegisterRoutes(mux *http.ServeMux)
}

// Service composes ServiceHealth, ServiceInfo, and Handler into the full
// contract that every microservice must satisfy. Embed this in your service
// type and implement all methods to satisfy the compiler.
type Service interface {
	ServiceHealth
	ServiceInfo
	Handler
}

// StandardError is the canonical JSON error body returned by all microservices.
// Every non-2xx HTTP response body MUST be serialisable to this type.
type StandardError struct {
	// Code is a machine-readable, SCREAMING_SNAKE_CASE identifier
	// (e.g. "NOT_FOUND", "VALIDATION_ERROR").
	Code string `json:"code"`
	// Message is a human-readable description intended for developers.
	Message string `json:"message"`
	// Details carries optional, structured context about the error.
	// Use this for field-level validation failures or additional metadata.
	Details map[string]any `json:"details,omitempty"`
}

// Error implements the error interface so StandardError can be used as an error.
func (e *StandardError) Error() string {
	if e.Message != "" {
		return e.Code + ": " + e.Message
	}
	return e.Code
}

// NewStandardError creates a StandardError with the given code and message.
func NewStandardError(code, message string) *StandardError {
	return &StandardError{Code: code, Message: message}
}

// NewStandardErrorWithDetails creates a StandardError with optional details.
func NewStandardErrorWithDetails(code, message string, details map[string]any) *StandardError {
	return &StandardError{Code: code, Message: message, Details: details}
}
