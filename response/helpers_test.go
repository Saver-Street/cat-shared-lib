package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"k": "v"})
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
}

func TestOK_Status200(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]int{"n": 1})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCreated_Status201(t *testing.T) {
	w := httptest.NewRecorder()
	Created(w, nil)
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
}

func TestError_BodyHasErrorKey(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad input")
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "bad input" {
		t.Errorf("error field = %q, want bad input", body["error"])
	}
}

func TestInternalError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "db fail", nil)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestDecodeOrFail_ValidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"x":1}`))
	w := httptest.NewRecorder()
	var v map[string]int
	if !DecodeOrFail(w, r, &v) {
		t.Error("expected true for valid JSON")
	}
}

func TestDecodeOrFail_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	w := httptest.NewRecorder()
	if DecodeOrFail(w, r, &struct{}{}) {
		t.Error("expected false for invalid JSON")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestJSON_EncodingError(t *testing.T) {
	w := httptest.NewRecorder()
	// Channels cannot be JSON-encoded — triggers the error path
	JSON(w, http.StatusOK, make(chan int))
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 (set before encode)", w.Code)
	}
}

func TestDecodeJSON_BodySizeLimit(t *testing.T) {
	// Generate a body larger than 1MB limit
	big := strings.Repeat("a", 1<<20+100)
	body := `{"data":"` + big + `"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var v map[string]string
	err := DecodeJSON(r, &v)
	if err == nil {
		t.Error("expected error for body exceeding 1MB limit")
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	var v map[string]string
	err := DecodeJSON(r, &v)
	if err == nil {
		t.Error("expected error for empty body")
	}
}

func TestDecodeJSON_ValidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
	var v map[string]string
	if err := DecodeJSON(r, &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v["name"] != "test" {
		t.Errorf("name = %q, want test", v["name"])
	}
}

func TestJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	body := strings.TrimSpace(w.Body.String())
	if body != "null" {
		t.Errorf("body = %q, want null", body)
	}
}

func TestOK_WithComplexData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{"items": []string{"a", "b"}, "count": 2}
	OK(w, data)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got["count"].(float64) != 2 {
		t.Errorf("count = %v, want 2", got["count"])
	}
}

func TestError_VariousStatusCodes(t *testing.T) {
	codes := []int{
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusConflict,
		http.StatusUnprocessableEntity,
	}
	for _, code := range codes {
		w := httptest.NewRecorder()
		Error(w, code, "msg")
		if w.Code != code {
			t.Errorf("Error(%d): got %d", code, w.Code)
		}
	}
}

func TestInternalError_WithRealError(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "operation failed", fmt.Errorf("connection refused"))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "Internal server error" {
		t.Errorf("error = %q, want Internal server error", body["error"])
	}
}

func TestDecodeOrFail_TruncatedJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"trun`))
	w := httptest.NewRecorder()
	if DecodeOrFail(w, r, &struct{ Name string }{}) {
		t.Error("expected false for truncated JSON")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestAccepted_Status202(t *testing.T) {
	w := httptest.NewRecorder()
	Accepted(w, map[string]string{"status": "queued"})
	if w.Code != http.StatusAccepted {
		t.Errorf("status = %d, want 202", w.Code)
	}
}

func TestNoContent_Status204(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
	if w.Body.Len() != 0 {
		t.Errorf("body should be empty, got %q", w.Body.String())
	}
}

func TestBadRequest_Status400(t *testing.T) {
	w := httptest.NewRecorder()
	BadRequest(w, "invalid input")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	if body["error"] != "invalid input" {
		t.Errorf("error = %q, want invalid input", body["error"])
	}
}

func TestUnauthorized_Status401(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "token expired")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestForbidden_Status403(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w, "access denied")
	if w.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", w.Code)
	}
}

func TestNotFound_Status404(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "resource not found")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestConflict_Status409(t *testing.T) {
	w := httptest.NewRecorder()
	Conflict(w, "already exists")
	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func TestUnprocessableEntity_Status422(t *testing.T) {
	w := httptest.NewRecorder()
	UnprocessableEntity(w, "validation failed")
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want 422", w.Code)
	}
}

// --- Benchmarks ---

func BenchmarkJSON(b *testing.B) {
	data := map[string]any{"id": "abc-123", "status": "ok", "count": 42}
	for b.Loop() {
		w := httptest.NewRecorder()
		JSON(w, http.StatusOK, data)
	}
}

func BenchmarkDecodeJSON(b *testing.B) {
	body := `{"name":"test","value":123}`
	for b.Loop() {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		var v map[string]any
		DecodeJSON(r, &v)
	}
}
