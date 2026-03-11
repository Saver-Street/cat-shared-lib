package response

import (
	"net/http"
	"time"
)

// HealthStatus represents the status of a health check.
type HealthStatus string

const (
	// StatusUp means the service is healthy.
	StatusUp HealthStatus = "up"
	// StatusDown means the service is unhealthy.
	StatusDown HealthStatus = "down"
	// StatusDegraded means the service is partially healthy.
	StatusDegraded HealthStatus = "degraded"
)

// ComponentHealth describes the health of an individual component.
type ComponentHealth struct {
	// Status is the component status.
	Status HealthStatus `json:"status"`
	// Details holds optional key-value metadata.
	Details map[string]any `json:"details,omitempty"`
}

// HealthResponse is the body of a health check endpoint.
type HealthResponse struct {
	// Status is the overall service status.
	Status HealthStatus `json:"status"`
	// Timestamp is when the check was performed.
	Timestamp string `json:"timestamp"`
	// Components maps component names to their health.
	Components map[string]ComponentHealth `json:"components,omitempty"`
}

// NewHealthResponse creates a HealthResponse with the given overall status.
func NewHealthResponse(status HealthStatus) HealthResponse {
	return HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// WithComponent adds a component to the response.
func (h HealthResponse) WithComponent(name string, ch ComponentHealth) HealthResponse {
	if h.Components == nil {
		h.Components = make(map[string]ComponentHealth)
	}
	h.Components[name] = ch
	return h
}

// Health writes a HealthResponse as JSON. It returns 200 for "up" or
// "degraded" and 503 for "down".
func Health(w http.ResponseWriter, h HealthResponse) {
	status := http.StatusOK
	if h.Status == StatusDown {
		status = http.StatusServiceUnavailable
	}
	JSON(w, status, h)
}
