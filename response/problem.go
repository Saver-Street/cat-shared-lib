package response

import (
	"encoding/json"
	"net/http"
)

// ProblemDetail represents an RFC 9457 (formerly RFC 7807) Problem Details
// object for conveying machine-readable error information in HTTP APIs.
type ProblemDetail struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type"`
	// Title is a short, human-readable summary of the problem type.
	Title string `json:"title"`
	// Status is the HTTP status code for this occurrence.
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference that identifies the specific occurrence.
	Instance string `json:"instance,omitempty"`
}

// WriteProblem writes a ProblemDetail as a JSON response with the
// application/problem+json content type and the appropriate status code.
func WriteProblem(w http.ResponseWriter, p ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// NewProblem creates a ProblemDetail with the given type, title, and status.
func NewProblem(typ, title string, status int) ProblemDetail {
	return ProblemDetail{Type: typ, Title: title, Status: status}
}

// WithDetail returns a copy of the ProblemDetail with the Detail field set.
func (p ProblemDetail) WithDetail(detail string) ProblemDetail {
	p.Detail = detail
	return p
}

// WithInstance returns a copy of the ProblemDetail with the Instance field set.
func (p ProblemDetail) WithInstance(instance string) ProblemDetail {
	p.Instance = instance
	return p
}

// Common pre-defined problem types following RFC 9457 conventions.
var (
	ProblemNotFound = NewProblem(
		"https://httpstatuses.io/404",
		"Not Found",
		http.StatusNotFound,
	)
	ProblemBadRequest = NewProblem(
		"https://httpstatuses.io/400",
		"Bad Request",
		http.StatusBadRequest,
	)
	ProblemUnauthorized = NewProblem(
		"https://httpstatuses.io/401",
		"Unauthorized",
		http.StatusUnauthorized,
	)
	ProblemForbidden = NewProblem(
		"https://httpstatuses.io/403",
		"Forbidden",
		http.StatusForbidden,
	)
	ProblemConflict = NewProblem(
		"https://httpstatuses.io/409",
		"Conflict",
		http.StatusConflict,
	)
	ProblemInternal = NewProblem(
		"https://httpstatuses.io/500",
		"Internal Server Error",
		http.StatusInternalServerError,
	)
)
