package response

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/validation"
)

func TestValidationErrors_WithValidationErrors(t *testing.T) {
	errs := []error{
		validation.Email("email", "bad"),
		validation.Required("name", ""),
	}

	w := httptest.NewRecorder()
	ValidationErrors(w, errs)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	var resp ValidationErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if resp.Error != "Validation failed" {
		t.Errorf("error = %q, want %q", resp.Error, "Validation failed")
	}
	if len(resp.Fields) != 2 {
		t.Fatalf("fields count = %d, want 2", len(resp.Fields))
	}
	if resp.Fields[0].Field != "email" {
		t.Errorf("field[0].field = %q, want %q", resp.Fields[0].Field, "email")
	}
	if resp.Fields[1].Field != "name" {
		t.Errorf("field[1].field = %q, want %q", resp.Fields[1].Field, "name")
	}
}

func TestValidationErrors_WithGenericErrors(t *testing.T) {
	errs := []error{
		errors.New("something went wrong"),
	}

	w := httptest.NewRecorder()
	ValidationErrors(w, errs)

	var resp ValidationErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}

	if len(resp.Fields) != 1 {
		t.Fatalf("fields count = %d, want 1", len(resp.Fields))
	}
	if resp.Fields[0].Field != "" {
		t.Errorf("field = %q, want empty", resp.Fields[0].Field)
	}
	if resp.Fields[0].Message != "something went wrong" {
		t.Errorf("message = %q, want %q", resp.Fields[0].Message, "something went wrong")
	}
}

func TestValidationErrors_Empty(t *testing.T) {
	w := httptest.NewRecorder()
	ValidationErrors(w, nil)

	var resp ValidationErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Fields) != 0 {
		t.Errorf("fields count = %d, want 0", len(resp.Fields))
	}
}

func TestValidationErrors_NilErrorsSkipped(t *testing.T) {
	errs := []error{
		nil,
		validation.Required("name", ""),
		nil,
	}

	w := httptest.NewRecorder()
	ValidationErrors(w, errs)

	var resp ValidationErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Fields) != 1 {
		t.Fatalf("fields count = %d, want 1", len(resp.Fields))
	}
	if resp.Fields[0].Field != "name" {
		t.Errorf("field = %q, want %q", resp.Fields[0].Field, "name")
	}
}

func TestValidationErrors_WithValidator(t *testing.T) {
	v := validation.NewValidator()
	v.Check(validation.Email("email", "not-an-email"))
	v.Check(validation.MinLength("password", "ab", 8))

	w := httptest.NewRecorder()
	ValidationErrors(w, v.Errors())

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	var resp ValidationErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Fields) != 2 {
		t.Fatalf("fields count = %d, want 2", len(resp.Fields))
	}
}

func TestValidationErrors_ContentType(t *testing.T) {
	w := httptest.NewRecorder()
	ValidationErrors(w, []error{validation.Required("x", "")})

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func BenchmarkValidationErrors(b *testing.B) {
	errs := []error{
		validation.Email("email", "bad"),
		validation.Required("name", ""),
		validation.MinLength("password", "x", 8),
	}
	for b.Loop() {
		w := httptest.NewRecorder()
		ValidationErrors(w, errs)
	}
}
