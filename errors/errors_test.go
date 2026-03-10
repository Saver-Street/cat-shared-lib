package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New(CodeNotFound, "user not found")
	if err.Code != CodeNotFound {
		t.Errorf("code = %q, want %q", err.Code, CodeNotFound)
	}
	if err.Message != "user not found" {
		t.Errorf("message = %q, want %q", err.Message, "user not found")
	}
	if err.Err != nil {
		t.Errorf("underlying err = %v, want nil", err.Err)
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("db: connection refused")
	err := Wrap(CodeInternal, "database error", cause)
	if err.Code != CodeInternal {
		t.Errorf("code = %q, want %q", err.Code, CodeInternal)
	}
	if err.Err != cause {
		t.Errorf("underlying err = %v, want %v", err.Err, cause)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is should match wrapped cause")
	}
}

func TestAppError_Error_WithUnderlying(t *testing.T) {
	cause := fmt.Errorf("timeout")
	err := Wrap(CodeGatewayTimeout, "upstream failed", cause)
	got := err.Error()
	want := "GATEWAY_TIMEOUT: upstream failed: timeout"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAppError_Error_NoUnderlying(t *testing.T) {
	err := New(CodeBadRequest, "invalid input")
	got := err.Error()
	want := "BAD_REQUEST: invalid input"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("root cause")
	err := Wrap(CodeInternal, "wrapped", cause)
	if err.Unwrap() != cause {
		t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
	}
}

func TestAppError_Unwrap_Nil(t *testing.T) {
	err := New(CodeBadRequest, "no cause")
	if err.Unwrap() != nil {
		t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
	}
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code Code
		want int
	}{
		{CodeBadRequest, http.StatusBadRequest},
		{CodeUnauthorized, http.StatusUnauthorized},
		{CodeForbidden, http.StatusForbidden},
		{CodeNotFound, http.StatusNotFound},
		{CodeConflict, http.StatusConflict},
		{CodeUnprocessableEntity, http.StatusUnprocessableEntity},
		{CodeTooManyRequests, http.StatusTooManyRequests},
		{CodeInternal, http.StatusInternalServerError},
		{CodeServiceUnavailable, http.StatusServiceUnavailable},
		{CodeGatewayTimeout, http.StatusGatewayTimeout},
		{CodeValidation, http.StatusUnprocessableEntity},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := New(tt.code, "test")
			if got := err.HTTPStatus(); got != tt.want {
				t.Errorf("HTTPStatus() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestAppError_HTTPStatus_Unknown(t *testing.T) {
	err := New(Code("UNKNOWN"), "test")
	if got := err.HTTPStatus(); got != http.StatusInternalServerError {
		t.Errorf("unknown code HTTPStatus() = %d, want 500", got)
	}
}

func TestHTTPStatusForCode(t *testing.T) {
	if got := HTTPStatusForCode(CodeNotFound); got != http.StatusNotFound {
		t.Errorf("HTTPStatusForCode(NOT_FOUND) = %d, want 404", got)
	}
	if got := HTTPStatusForCode(Code("UNKNOWN")); got != http.StatusInternalServerError {
		t.Errorf("HTTPStatusForCode(UNKNOWN) = %d, want 500", got)
	}
}

func TestConvenienceConstructors(t *testing.T) {
	tests := []struct {
		name string
		fn   func(string) *AppError
		code Code
	}{
		{"BadRequest", BadRequest, CodeBadRequest},
		{"Unauthorized", Unauthorized, CodeUnauthorized},
		{"Forbidden", Forbidden, CodeForbidden},
		{"NotFound", NotFound, CodeNotFound},
		{"Conflict", Conflict, CodeConflict},
		{"Internal", Internal, CodeInternal},
		{"Validation", Validation, CodeValidation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn("test message")
			if err.Code != tt.code {
				t.Errorf("code = %q, want %q", err.Code, tt.code)
			}
			if err.Message != "test message" {
				t.Errorf("message = %q, want %q", err.Message, "test message")
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	err := NotFound("not found")
	var appErr *AppError
	if !errors.As(err, &appErr) {
		t.Error("errors.As should match *AppError")
	}
	if appErr.Code != CodeNotFound {
		t.Errorf("code = %q, want %q", appErr.Code, CodeNotFound)
	}
}

func TestWrappedErrorsAs(t *testing.T) {
	cause := fmt.Errorf("root: %w", Validation("bad email"))
	var appErr *AppError
	if !errors.As(cause, &appErr) {
		t.Error("errors.As should match wrapped *AppError")
	}
}

func BenchmarkNew(b *testing.B) {
	for b.Loop() {
		New(CodeNotFound, "not found")
	}
}

func BenchmarkWrap(b *testing.B) {
	cause := fmt.Errorf("timeout")
	for b.Loop() {
		Wrap(CodeInternal, "db error", cause)
	}
}

func BenchmarkHTTPStatus(b *testing.B) {
	err := New(CodeNotFound, "not found")
	for b.Loop() {
		err.HTTPStatus()
	}
}
