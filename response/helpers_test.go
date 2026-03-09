package response

import (
	"encoding/json"
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
