package apperror

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	e := New(http.StatusNotFound, CodeNotFound, "user not found")
	if e.HTTPStatus != 404 {
		t.Errorf("HTTPStatus = %d, want 404", e.HTTPStatus)
	}
	if e.Code != CodeNotFound {
		t.Errorf("Code = %q, want %q", e.Code, CodeNotFound)
	}
	if e.Message != "user not found" {
		t.Errorf("Message = %q, want %q", e.Message, "user not found")
	}
	if e.Err != nil {
		t.Errorf("Err = %v, want nil", e.Err)
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("db connection lost")
	e := Wrap(http.StatusInternalServerError, CodeInternal, "failed to fetch user", cause)
	if e.Err != cause {
		t.Errorf("Err = %v, want %v", e.Err, cause)
	}
	if !errors.Is(e, cause) {
		t.Error("errors.Is should find wrapped cause")
	}
}

func TestError_Error_WithCause(t *testing.T) {
	cause := fmt.Errorf("timeout")
	e := Wrap(500, CodeInternal, "query failed", cause)
	got := e.Error()
	want := "INTERNAL_ERROR: query failed: timeout"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_Error_NoCause(t *testing.T) {
	e := New(404, CodeNotFound, "not found")
	got := e.Error()
	want := "NOT_FOUND: not found"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original")
	e := Wrap(500, CodeInternal, "wrapped", cause)
	if e.Unwrap() != cause {
		t.Error("Unwrap should return cause")
	}

	e2 := New(400, CodeBadRequest, "no cause")
	if e2.Unwrap() != nil {
		t.Error("Unwrap should return nil when no cause")
	}
}

func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(string) *Error
		wantStatus int
		wantCode   Code
	}{
		{"NotFound", NotFound, 404, CodeNotFound},
		{"BadRequest", BadRequest, 400, CodeBadRequest},
		{"Unauthorized", Unauthorized, 401, CodeUnauthorized},
		{"Forbidden", Forbidden, 403, CodeForbidden},
		{"Conflict", Conflict, 409, CodeConflict},
		{"Validation", Validation, 422, CodeValidation},
		{"Internal", Internal, 500, CodeInternal},
		{"Timeout", Timeout, 504, CodeTimeout},
		{"RateLimit", RateLimit, 429, CodeRateLimit},
		{"ServiceDown", ServiceDown, 503, CodeServiceDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fn("test message")
			if e.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, tt.wantStatus)
			}
			if e.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tt.wantCode)
			}
			if e.Message != "test message" {
				t.Errorf("Message = %q", e.Message)
			}
		})
	}
}

func TestInternalWrap(t *testing.T) {
	cause := fmt.Errorf("disk full")
	e := InternalWrap("write failed", cause)
	if e.HTTPStatus != 500 {
		t.Errorf("HTTPStatus = %d, want 500", e.HTTPStatus)
	}
	if e.Code != CodeInternal {
		t.Errorf("Code = %q, want %q", e.Code, CodeInternal)
	}
	if !errors.Is(e, cause) {
		t.Error("should wrap cause")
	}
}

func TestHTTPStatus(t *testing.T) {
	if got := HTTPStatus(nil); got != 200 {
		t.Errorf("HTTPStatus(nil) = %d, want 200", got)
	}
	if got := HTTPStatus(NotFound("nope")); got != 404 {
		t.Errorf("HTTPStatus(NotFound) = %d, want 404", got)
	}
	if got := HTTPStatus(fmt.Errorf("random")); got != 500 {
		t.Errorf("HTTPStatus(random error) = %d, want 500", got)
	}
}

func TestIsCode(t *testing.T) {
	e := NotFound("user")
	if !IsCode(e, CodeNotFound) {
		t.Error("IsCode should match")
	}
	if IsCode(e, CodeBadRequest) {
		t.Error("IsCode should not match wrong code")
	}
	if IsCode(fmt.Errorf("plain error"), CodeNotFound) {
		t.Error("IsCode should not match plain error")
	}
}
