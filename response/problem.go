package response

import (
	"encoding/json"
	"net/http"
)

// ProblemDetail represents an RFC 7807 problem detail response.
// See https://tools.ietf.org/html/rfc7807
type ProblemDetail struct {
	// Type is a URI reference that identifies the problem type.
	Type string `json:"type"`
	// Title is a short, human-readable summary.
	Title string `json:"title"`
	// Status is the HTTP status code.
	Status int `json:"status"`
	// Detail is a human-readable explanation specific to this occurrence.
	Detail string `json:"detail,omitempty"`
	// Instance is a URI reference for the specific occurrence.
	Instance string `json:"instance,omitempty"`
	// Extensions holds any additional members.
	Extensions map[string]any `json:"extensions,omitempty"`
}

// Problem writes a ProblemDetail as an application/problem+json response.
func Problem(w http.ResponseWriter, p ProblemDetail) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(p.Status)
	_ = json.NewEncoder(w).Encode(p)
}

// NewProblem creates a ProblemDetail with the given type, title, and status.
func NewProblem(typ, title string, status int) ProblemDetail {
	return ProblemDetail{
		Type:   typ,
		Title:  title,
		Status: status,
	}
}

// WithDetail returns a copy of the ProblemDetail with the detail field set.
func (p ProblemDetail) WithDetail(detail string) ProblemDetail {
	p.Detail = detail
	return p
}

// WithInstance returns a copy with the instance URI set.
func (p ProblemDetail) WithInstance(instance string) ProblemDetail {
	p.Instance = instance
	return p
}

// WithExtension returns a copy with an additional extension member.
func (p ProblemDetail) WithExtension(key string, value any) ProblemDetail {
	if p.Extensions == nil {
		p.Extensions = make(map[string]any)
	}
	p.Extensions[key] = value
	return p
}
