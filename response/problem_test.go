package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProblem(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	p := NewProblem("about:blank", "Not Found", http.StatusNotFound)
	Problem(w, p)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d; want %d", w.Code, http.StatusNotFound)
	}
	ct := w.Header().Get("Content-Type")
	if ct != "application/problem+json" {
		t.Errorf("Content-Type = %q; want application/problem+json", ct)
	}

	var got ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != "about:blank" {
		t.Errorf("Type = %q; want about:blank", got.Type)
	}
	if got.Title != "Not Found" {
		t.Errorf("Title = %q; want Not Found", got.Title)
	}
	if got.Status != 404 {
		t.Errorf("Status = %d; want 404", got.Status)
	}
}

func TestProblemWithDetail(t *testing.T) {
	t.Parallel()
	p := NewProblem("urn:err:balance", "Insufficient Funds", 403).
		WithDetail("Your account balance is too low.")

	w := httptest.NewRecorder()
	Problem(w, p)

	var got ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Detail != "Your account balance is too low." {
		t.Errorf("Detail = %q", got.Detail)
	}
}

func TestProblemWithInstance(t *testing.T) {
	t.Parallel()
	p := NewProblem("about:blank", "Error", 500).
		WithInstance("/requests/abc-123")

	w := httptest.NewRecorder()
	Problem(w, p)

	var got ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Instance != "/requests/abc-123" {
		t.Errorf("Instance = %q", got.Instance)
	}
}

func TestProblemWithExtension(t *testing.T) {
	t.Parallel()
	p := NewProblem("urn:err:rate", "Rate Limit", 429).
		WithExtension("retryAfter", 30).
		WithExtension("limit", 100)

	w := httptest.NewRecorder()
	Problem(w, p)

	var got ProblemDetail
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Extensions == nil {
		t.Fatal("Extensions is nil")
	}
	if v, ok := got.Extensions["retryAfter"]; !ok || v != float64(30) {
		t.Errorf("retryAfter = %v", v)
	}
	if v, ok := got.Extensions["limit"]; !ok || v != float64(100) {
		t.Errorf("limit = %v", v)
	}
}

func TestProblemOmitsEmpty(t *testing.T) {
	t.Parallel()
	p := NewProblem("about:blank", "OK", 200)
	w := httptest.NewRecorder()
	Problem(w, p)

	var raw map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := raw["detail"]; ok {
		t.Error("empty detail should be omitted")
	}
	if _, ok := raw["instance"]; ok {
		t.Error("empty instance should be omitted")
	}
	if _, ok := raw["extensions"]; ok {
		t.Error("nil extensions should be omitted")
	}
}

func TestNewProblem(t *testing.T) {
	t.Parallel()
	p := NewProblem("urn:x", "T", 400)
	if p.Type != "urn:x" || p.Title != "T" || p.Status != 400 {
		t.Errorf("NewProblem = %+v", p)
	}
}

func TestProblemChaining(t *testing.T) {
	t.Parallel()
	p := NewProblem("urn:x", "T", 400).
		WithDetail("d").
		WithInstance("i").
		WithExtension("k", "v")
	if p.Detail != "d" || p.Instance != "i" {
		t.Errorf("chaining failed: %+v", p)
	}
	if p.Extensions["k"] != "v" {
		t.Errorf("extension = %v", p.Extensions["k"])
	}
}

func BenchmarkProblem(b *testing.B) {
	p := NewProblem("urn:err:test", "Test", 400).WithDetail("d")
	for range b.N {
		w := httptest.NewRecorder()
		Problem(w, p)
	}
}

func BenchmarkNewProblem(b *testing.B) {
	for range b.N {
		_ = NewProblem("urn:err:test", "Test", 400)
	}
}

func FuzzProblemDetail(f *testing.F) {
	f.Add("about:blank", "Not Found", 404, "detail")
	f.Add("", "", 200, "")
	f.Fuzz(func(t *testing.T, typ, title string, status int, detail string) {
		if status < 100 || status > 999 {
			return
		}
		p := NewProblem(typ, title, status).WithDetail(detail)
		w := httptest.NewRecorder()
		Problem(w, p)
		if w.Code != status {
			t.Errorf("status = %d; want %d", w.Code, status)
		}
	})
}
