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

func TestTooManyRequests_Status429(t *testing.T) {
	w := httptest.NewRecorder()
	TooManyRequests(w, "rate limit exceeded")
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", w.Code)
	}
}

func TestServiceUnavailable_Status503(t *testing.T) {
	w := httptest.NewRecorder()
	ServiceUnavailable(w, "service down")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", w.Code)
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

func BenchmarkOK(b *testing.B) {
	data := map[string]any{"id": "abc-123", "status": "ok"}
	for b.Loop() {
		w := httptest.NewRecorder()
		OK(w, data)
	}
}

func BenchmarkCreated(b *testing.B) {
	data := map[string]string{"id": "new-123"}
	for b.Loop() {
		w := httptest.NewRecorder()
		Created(w, data)
	}
}

func BenchmarkError(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		Error(w, http.StatusBadRequest, "bad input")
	}
}

func BenchmarkNoContent(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		NoContent(w)
	}
}

func BenchmarkBadRequest(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		BadRequest(w, "invalid input")
	}
}

func BenchmarkInternalError(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		InternalError(w, "db fail", nil)
	}
}

func BenchmarkAccepted(b *testing.B) {
	data := map[string]string{"status": "queued"}
	for b.Loop() {
		w := httptest.NewRecorder()
		Accepted(w, data)
	}
}

func BenchmarkNotFound(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		NotFound(w, "resource not found")
	}
}

func BenchmarkConflict(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		Conflict(w, "already exists")
	}
}

func BenchmarkForbidden(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		Forbidden(w, "access denied")
	}
}

func BenchmarkUnauthorized(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		Unauthorized(w, "token expired")
	}
}

func BenchmarkUnprocessableEntity(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		UnprocessableEntity(w, "validation failed")
	}
}

func BenchmarkTooManyRequests(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		TooManyRequests(w, "rate limit exceeded")
	}
}

func BenchmarkServiceUnavailable(b *testing.B) {
	for b.Loop() {
		w := httptest.NewRecorder()
		ServiceUnavailable(w, "service down")
	}
}

func BenchmarkDecodeOrFail(b *testing.B) {
	body := `{"name":"test","value":123}`
	for b.Loop() {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		w := httptest.NewRecorder()
		var v map[string]any
		DecodeOrFail(w, r, &v)
	}
}

func TestMethodNotAllowed_Status405(t *testing.T) {
w := httptest.NewRecorder()
MethodNotAllowed(w, "method not allowed")
if w.Code != http.StatusMethodNotAllowed {
t.Errorf("expected 405, got %d", w.Code)
}
}

func TestGone_Status410(t *testing.T) {
w := httptest.NewRecorder()
Gone(w, "resource deleted")
if w.Code != http.StatusGone {
t.Errorf("expected 410, got %d", w.Code)
}
}

func TestGatewayTimeout_Status504(t *testing.T) {
w := httptest.NewRecorder()
GatewayTimeout(w, "upstream timeout")
if w.Code != http.StatusGatewayTimeout {
t.Errorf("expected 504, got %d", w.Code)
}
}

func TestPaginated_EnvelopeFields(t *testing.T) {
w := httptest.NewRecorder()
Paginated(w, []string{"a", "b", "c"}, 50, 2, 3)

if w.Code != http.StatusOK {
t.Fatalf("expected 200, got %d", w.Code)
}
var got PagedResult[string]
if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
t.Fatalf("decode: %v", err)
}
if got.Total != 50 || got.Page != 2 || got.Limit != 3 || len(got.Data) != 3 {
t.Errorf("fields wrong: %+v", got)
}
if !got.HasMore {
t.Error("has_more should be true (page*limit=6 < total=50)")
}
}

func TestPaginated_HasMore_False_OnLastPage(t *testing.T) {
w := httptest.NewRecorder()
Paginated(w, []int{4, 5}, 5, 2, 3)

var got PagedResult[int]
if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
t.Fatalf("decode: %v", err)
}
if got.HasMore {
t.Error("has_more should be false (page*limit=6 >= total=5)")
}
}

func TestPaginated_EmptyData(t *testing.T) {
w := httptest.NewRecorder()
Paginated(w, []string{}, 0, 1, 10)

var got PagedResult[string]
if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
t.Fatalf("decode: %v", err)
}
if got.Total != 0 || got.HasMore {
t.Errorf("empty result wrong: %+v", got)
}
}
