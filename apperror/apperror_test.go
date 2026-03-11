package apperror

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNew(t *testing.T) {
	e := New(http.StatusNotFound, CodeNotFound, "user not found")
	testkit.AssertEqual(t, e.HTTPStatus, 404)
	testkit.AssertEqual(t, e.Code, CodeNotFound)
	testkit.AssertEqual(t, e.Message, "user not found")
	testkit.AssertNil(t, e.Err)
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("db connection lost")
	e := Wrap(http.StatusInternalServerError, CodeInternal, "failed to fetch user", cause)
	testkit.AssertErrorIs(t, e, cause)
	testkit.AssertErrorIs(t, e, cause)
}

func TestError_Error_WithCause(t *testing.T) {
	cause := fmt.Errorf("timeout")
	e := Wrap(500, CodeInternal, "query failed", cause)
	got := e.Error()
	want := "INTERNAL_ERROR: query failed: timeout"
	testkit.AssertEqual(t, got, want)
}

func TestError_Error_NoCause(t *testing.T) {
	e := New(404, CodeNotFound, "not found")
	got := e.Error()
	want := "NOT_FOUND: not found"
	testkit.AssertEqual(t, got, want)
}

func TestError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original")
	e := Wrap(500, CodeInternal, "wrapped", cause)
	testkit.AssertErrorIs(t, e, cause)
	testkit.AssertErrorIs(t, e, cause)

	e2 := New(400, CodeBadRequest, "no cause")
	testkit.AssertNil(t, e2.Unwrap())
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
		{"PaymentRequired", PaymentRequired, 402, CodePaymentRequired},
		{"TooLarge", TooLarge, 413, CodeTooLarge},
		{"Gone", Gone, 410, CodeGone},
		{"NotImplemented", NotImplemented, 501, CodeNotImplemented},
		{"PreconditionFailed", PreconditionFailed, 412, CodePrecondition},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fn("test message")
			testkit.AssertEqual(t, e.HTTPStatus, tt.wantStatus)
			testkit.AssertEqual(t, e.Code, tt.wantCode)
			testkit.AssertEqual(t, e.Message, "test message")
		})
	}
}

func TestInternalWrap(t *testing.T) {
	cause := fmt.Errorf("disk full")
	e := InternalWrap("write failed", cause)
	testkit.AssertEqual(t, e.HTTPStatus, 500)
	testkit.AssertEqual(t, e.Code, CodeInternal)
	testkit.AssertErrorIs(t, e, cause)
}

func TestWrapVariants(t *testing.T) {
	cause := fmt.Errorf("root cause")

	tests := []struct {
		name       string
		fn         func(string, error) *Error
		wantStatus int
		wantCode   Code
	}{
		{"NotFoundWrap", NotFoundWrap, 404, CodeNotFound},
		{"BadRequestWrap", BadRequestWrap, 400, CodeBadRequest},
		{"UnauthorizedWrap", UnauthorizedWrap, 401, CodeUnauthorized},
		{"ForbiddenWrap", ForbiddenWrap, 403, CodeForbidden},
		{"ConflictWrap", ConflictWrap, 409, CodeConflict},
		{"ValidationWrap", ValidationWrap, 422, CodeValidation},
		{"TimeoutWrap", TimeoutWrap, 504, CodeTimeout},
		{"ServiceDownWrap", ServiceDownWrap, 503, CodeServiceDown},
		{"InternalWrap", InternalWrap, 500, CodeInternal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := tt.fn("wrap message", cause)
			testkit.AssertEqual(t, e.HTTPStatus, tt.wantStatus)
			testkit.AssertEqual(t, e.Code, tt.wantCode)
			testkit.AssertEqual(t, e.Message, "wrap message")
			testkit.AssertErrorIs(t, e, cause)
		})
	}
}

func TestHTTPStatus(t *testing.T) {
	testkit.AssertEqual(t, HTTPStatus(nil), 200)
	testkit.AssertEqual(t, HTTPStatus(NotFound("nope")), 404)
	testkit.AssertEqual(t, HTTPStatus(fmt.Errorf("random")), 500)
}

func TestIsCode(t *testing.T) {
	e := NotFound("user")
	testkit.AssertTrue(t, IsCode(e, CodeNotFound))
	testkit.AssertFalse(t, IsCode(e, CodeBadRequest))
	testkit.AssertFalse(t, IsCode(fmt.Errorf("plain error"), CodeNotFound))
}
