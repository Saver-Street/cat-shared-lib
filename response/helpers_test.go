package response

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/apperror"
	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestJSON_SetsContentType(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, map[string]string{"k": "v"})
	testkit.AssertHeader(t, w, "Content-Type", "application/json")
}

func TestOK_Status200(t *testing.T) {
	w := httptest.NewRecorder()
	OK(w, map[string]int{"n": 1})
	testkit.AssertStatus(t, w, http.StatusOK)
}

func TestCreated_Status201(t *testing.T) {
	w := httptest.NewRecorder()
	Created(w, nil)
	testkit.AssertStatus(t, w, http.StatusCreated)
}

func TestError_BodyHasErrorKey(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad input")
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["error"], "bad input")
}

func TestInternalError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "db fail", nil)
	testkit.AssertStatus(t, w, http.StatusInternalServerError)
}

func TestDecodeOrFail_ValidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"x":1}`))
	w := httptest.NewRecorder()
	var v map[string]int
	testkit.AssertTrue(t, DecodeOrFail(w, r, &v))
}

func TestDecodeOrFail_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{bad"))
	w := httptest.NewRecorder()
	testkit.AssertFalse(t, DecodeOrFail(w, r, &struct{}{}))
	testkit.AssertStatus(t, w, http.StatusBadRequest)
}

func TestJSON_EncodingError(t *testing.T) {
	w := httptest.NewRecorder()
	// Channels cannot be JSON-encoded — triggers the error path
	JSON(w, http.StatusOK, make(chan int))
	testkit.AssertStatus(t, w, http.StatusOK)
}

func TestDecodeJSON_BodySizeLimit(t *testing.T) {
	// Generate a body larger than 1MB limit
	big := strings.Repeat("a", 1<<20+100)
	body := `{"data":"` + big + `"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var v map[string]string
	err := DecodeJSON(r, &v)
	testkit.AssertError(t, err)
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	var v map[string]string
	err := DecodeJSON(r, &v)
	testkit.AssertError(t, err)
}

func TestDecodeJSON_ValidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"test"}`))
	var v map[string]string
	if err := DecodeJSON(r, &v); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	testkit.AssertEqual(t, v["name"], "test")
}

func TestJSON_NilData(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusOK, nil)
	testkit.AssertStatus(t, w, http.StatusOK)
	body := strings.TrimSpace(w.Body.String())
	testkit.AssertEqual(t, body, "null")
}

func TestOK_WithComplexData(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]any{"items": []string{"a", "b"}, "count": 2}
	OK(w, data)
	testkit.AssertStatus(t, w, http.StatusOK)
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	testkit.AssertEqual(t, got["count"].(float64), float64(2))
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
		testkit.AssertStatus(t, w, code)
	}
}

func TestInternalError_WithRealError(t *testing.T) {
	w := httptest.NewRecorder()
	InternalError(w, "operation failed", fmt.Errorf("connection refused"))
	testkit.AssertStatus(t, w, http.StatusInternalServerError)
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["error"], "Internal server error")
}

func TestDecodeOrFail_TruncatedJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"trun`))
	w := httptest.NewRecorder()
	testkit.AssertFalse(t, DecodeOrFail(w, r, &struct{ Name string }{}))
	testkit.AssertStatus(t, w, http.StatusBadRequest)
}

func TestAccepted_Status202(t *testing.T) {
	w := httptest.NewRecorder()
	Accepted(w, map[string]string{"status": "queued"})
	testkit.AssertStatus(t, w, http.StatusAccepted)
}

func TestNoContent_Status204(t *testing.T) {
	w := httptest.NewRecorder()
	NoContent(w)
	testkit.AssertStatus(t, w, http.StatusNoContent)
	testkit.AssertEqual(t, w.Body.Len(), 0)
}

func TestBadRequest_Status400(t *testing.T) {
	w := httptest.NewRecorder()
	BadRequest(w, "invalid input")
	testkit.AssertStatus(t, w, http.StatusBadRequest)
	var body map[string]string
	json.Unmarshal(w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["error"], "invalid input")
}

func TestUnauthorized_Status401(t *testing.T) {
	w := httptest.NewRecorder()
	Unauthorized(w, "token expired")
	testkit.AssertStatus(t, w, http.StatusUnauthorized)
}

func TestForbidden_Status403(t *testing.T) {
	w := httptest.NewRecorder()
	Forbidden(w, "access denied")
	testkit.AssertStatus(t, w, http.StatusForbidden)
}

func TestNotFound_Status404(t *testing.T) {
	w := httptest.NewRecorder()
	NotFound(w, "resource not found")
	testkit.AssertStatus(t, w, http.StatusNotFound)
}

func TestConflict_Status409(t *testing.T) {
	w := httptest.NewRecorder()
	Conflict(w, "already exists")
	testkit.AssertStatus(t, w, http.StatusConflict)
}

func TestUnprocessableEntity_Status422(t *testing.T) {
	w := httptest.NewRecorder()
	UnprocessableEntity(w, "validation failed")
	testkit.AssertStatus(t, w, http.StatusUnprocessableEntity)
}

func TestTooManyRequests_Status429(t *testing.T) {
	w := httptest.NewRecorder()
	TooManyRequests(w, "rate limit exceeded")
	testkit.AssertStatus(t, w, http.StatusTooManyRequests)
}

func TestServiceUnavailable_Status503(t *testing.T) {
	w := httptest.NewRecorder()
	ServiceUnavailable(w, "service down")
	testkit.AssertStatus(t, w, http.StatusServiceUnavailable)
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
	testkit.AssertStatus(t, w, http.StatusMethodNotAllowed)
}

func TestGone_Status410(t *testing.T) {
	w := httptest.NewRecorder()
	Gone(w, "resource deleted")
	testkit.AssertStatus(t, w, http.StatusGone)
}

func TestGatewayTimeout_Status504(t *testing.T) {
	w := httptest.NewRecorder()
	GatewayTimeout(w, "upstream timeout")
	testkit.AssertStatus(t, w, http.StatusGatewayTimeout)
}

func TestPaginated_EnvelopeFields(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{"a", "b", "c"}, 50, 2, 3)

		testkit.RequireEqual(t, w.Code, http.StatusOK)
	var got PagedResult[string]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertEqual(t, got.Total, 50)
	testkit.AssertEqual(t, got.Page, 2)
	testkit.AssertEqual(t, got.Limit, 3)
	testkit.AssertLen(t, got.Data, 3)
	testkit.AssertTrue(t, got.HasMore)
}

func TestPaginated_HasMore_False_OnLastPage(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []int{4, 5}, 5, 2, 3)

	var got PagedResult[int]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertFalse(t, got.HasMore)
}

func TestPaginated_EmptyData(t *testing.T) {
	w := httptest.NewRecorder()
	Paginated(w, []string{}, 0, 1, 10)

	var got PagedResult[string]
	testkit.AssertJSON(t, w.Body.Bytes(), &got)
	testkit.AssertEqual(t, got.Total, 0)
	testkit.AssertFalse(t, got.HasMore)
}

func TestAppError_WithAppError(t *testing.T) {
	w := httptest.NewRecorder()
	AppError(w, apperror.NotFound("user not found"))

	testkit.AssertStatus(t, w, http.StatusNotFound)
	testkit.AssertHeader(t, w, "Content-Type", "application/json")

	var body map[string]any
	testkit.AssertJSON(t, w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["code"], "NOT_FOUND")
	testkit.AssertEqual(t, body["message"], "user not found")
}

func TestAppError_WithWrappedAppError(t *testing.T) {
	inner := apperror.BadRequest("invalid input")
	wrapped := fmt.Errorf("handler: %w", inner)

	w := httptest.NewRecorder()
	AppError(w, wrapped)

	testkit.AssertStatus(t, w, http.StatusBadRequest)
}

func TestAppError_WithGenericError(t *testing.T) {
	w := httptest.NewRecorder()
	AppError(w, errors.New("something broke"))

	testkit.AssertStatus(t, w, http.StatusInternalServerError)

	var body map[string]string
	testkit.AssertJSON(t, w.Body.Bytes(), &body)
	testkit.AssertEqual(t, body["error"], "Internal server error")
}

func TestRedirect_Found(t *testing.T) {
	w := httptest.NewRecorder()
	r := testkit.NewRequest("GET", "/old", nil)
	Redirect(w, r, "/new", http.StatusFound)
	testkit.AssertEqual(t, w.Code, http.StatusFound)
	testkit.AssertEqual(t, w.Header().Get("Location"), "/new")
}

func TestRedirect_MovedPermanently(t *testing.T) {
	w := httptest.NewRecorder()
	r := testkit.NewRequest("GET", "/legacy", nil)
	Redirect(w, r, "https://example.com/new", http.StatusMovedPermanently)
	testkit.AssertEqual(t, w.Code, http.StatusMovedPermanently)
	testkit.AssertEqual(t, w.Header().Get("Location"), "https://example.com/new")
}

func TestText(t *testing.T) {
	w := httptest.NewRecorder()
	Text(w, http.StatusOK, "hello world")
	testkit.AssertEqual(t, w.Code, http.StatusOK)
	testkit.AssertEqual(t, w.Header().Get("Content-Type"), "text/plain; charset=utf-8")
	testkit.AssertEqual(t, w.Body.String(), "hello world")
}

func TestText_CustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	Text(w, http.StatusAccepted, "processing")
	testkit.AssertEqual(t, w.Code, http.StatusAccepted)
	testkit.AssertEqual(t, w.Body.String(), "processing")
}
