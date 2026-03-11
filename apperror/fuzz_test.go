package apperror

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func FuzzNew(f *testing.F) {
	f.Add(400, "BAD_REQUEST", "invalid input")
	f.Add(500, "INTERNAL_ERROR", "")
	f.Add(0, "", "")
	f.Add(999, "CUSTOM", "unicode: 日本語 ñ ü 🎉")
	f.Fuzz(func(t *testing.T, status int, code, message string) {
		e := New(status, Code(code), message)
		testkit.RequireNotNil(t, e)
		if e.HTTPStatus != status {
			t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, status)
		}
		// Error() must not panic.
		_ = e.Error()
		// HTTPStatus must return the status code.
		if got := HTTPStatus(e); got != status {
			t.Errorf("HTTPStatus(e) = %d, want %d", got, status)
		}
		// IsCode must match.
		if !IsCode(e, Code(code)) {
			t.Error("IsCode returned false for matching code")
		}
	})
}

func FuzzWrap(f *testing.F) {
	f.Add(500, "INTERNAL_ERROR", "something broke")
	f.Add(404, "NOT_FOUND", "")
	f.Fuzz(func(t *testing.T, status int, code, message string) {
		inner := errors.New("root cause")
		e := Wrap(status, Code(code), message, inner)
		testkit.RequireNotNil(t, e)
		// Error() must not panic and must include the inner error.
		s := e.Error()
		if s == "" {
			t.Error("Error() returned empty string")
		}
		// Unwrap must return inner error.
		if !errors.Is(e, inner) {
			t.Error("errors.Is failed for wrapped error")
		}
	})
}

func FuzzHTTPStatus(f *testing.F) {
	f.Add(200, "OK", "success")
	f.Add(0, "", "")
	f.Add(-1, "NEG", "negative status")
	f.Fuzz(func(t *testing.T, status int, code, message string) {
		e := New(status, Code(code), message)
		got := HTTPStatus(e)
		if got != status {
			t.Errorf("HTTPStatus = %d, want %d", got, status)
		}
		// Plain error always returns 500.
		plain := fmt.Errorf("plain error")
		if got := HTTPStatus(plain); got != 500 {
			t.Errorf("HTTPStatus(plain) = %d, want 500", got)
		}
	})
}
