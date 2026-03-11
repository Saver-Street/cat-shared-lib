package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteProblem(t *testing.T) {
	w := httptest.NewRecorder()
	p := ProblemDetail{
		Type:   "https://example.com/not-found",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: "User 123 was not found",
	}
	WriteProblem(w, p)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Fatalf("Content-Type = %q, want application/problem+json", ct)
	}

	var got ProblemDetail
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Type != p.Type {
		t.Fatalf("type = %q, want %q", got.Type, p.Type)
	}
	if got.Title != p.Title {
		t.Fatalf("title = %q, want %q", got.Title, p.Title)
	}
	if got.Status != p.Status {
		t.Fatalf("status = %d, want %d", got.Status, p.Status)
	}
	if got.Detail != p.Detail {
		t.Fatalf("detail = %q, want %q", got.Detail, p.Detail)
	}
}

func TestWriteProblemNoDetail(t *testing.T) {
	w := httptest.NewRecorder()
	p := ProblemDetail{
		Type:   "https://example.com/bad",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
	}
	WriteProblem(w, p)

	var m map[string]any
	if err := json.NewDecoder(w.Body).Decode(&m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["detail"]; ok {
		t.Fatal("detail should be omitted when empty")
	}
	if _, ok := m["instance"]; ok {
		t.Fatal("instance should be omitted when empty")
	}
}

func TestNewProblem(t *testing.T) {
	p := NewProblem("https://example.com/err", "Error", 500)
	if p.Type != "https://example.com/err" {
		t.Fatalf("type = %q", p.Type)
	}
	if p.Title != "Error" {
		t.Fatalf("title = %q", p.Title)
	}
	if p.Status != 500 {
		t.Fatalf("status = %d", p.Status)
	}
}

func TestWithDetail(t *testing.T) {
	p := NewProblem("t", "T", 400).WithDetail("details here")
	if p.Detail != "details here" {
		t.Fatalf("detail = %q", p.Detail)
	}
}

func TestWithInstance(t *testing.T) {
	p := NewProblem("t", "T", 400).WithInstance("/errors/123")
	if p.Instance != "/errors/123" {
		t.Fatalf("instance = %q", p.Instance)
	}
}

func TestWithDetailChained(t *testing.T) {
	p := ProblemNotFound.
		WithDetail("User not found").
		WithInstance("/users/123")
	if p.Detail != "User not found" {
		t.Fatalf("detail = %q", p.Detail)
	}
	if p.Instance != "/users/123" {
		t.Fatalf("instance = %q", p.Instance)
	}
	// Original should be unchanged
	if ProblemNotFound.Detail != "" {
		t.Fatal("original ProblemNotFound should be unchanged")
	}
}

func TestPredefinedProblems(t *testing.T) {
	tests := []struct {
		name   string
		p      ProblemDetail
		status int
	}{
		{"NotFound", ProblemNotFound, 404},
		{"BadRequest", ProblemBadRequest, 400},
		{"Unauthorized", ProblemUnauthorized, 401},
		{"Forbidden", ProblemForbidden, 403},
		{"Conflict", ProblemConflict, 409},
		{"Internal", ProblemInternal, 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.p.Status != tt.status {
				t.Fatalf("status = %d, want %d", tt.p.Status, tt.status)
			}
			if tt.p.Type == "" {
				t.Fatal("type should not be empty")
			}
			if tt.p.Title == "" {
				t.Fatal("title should not be empty")
			}
		})
	}
}

func BenchmarkWriteProblem(b *testing.B) {
	p := ProblemNotFound.WithDetail("not found")
	for b.Loop() {
		w := httptest.NewRecorder()
		WriteProblem(w, p)
	}
}

func BenchmarkNewProblem(b *testing.B) {
	for b.Loop() {
		NewProblem("type", "title", 400).
			WithDetail("detail").
			WithInstance("/inst")
	}
}
