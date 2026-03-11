package apperror_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/apperror"
)

func ExampleNew() {
	err := apperror.New(http.StatusConflict, apperror.CodeConflict, "user already exists")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	fmt.Println(err.Error())
	// Output:
	// 409
	// CONFLICT
	// CONFLICT: user already exists
}

func ExampleWrap() {
	original := errors.New("connection refused")
	err := apperror.Wrap(http.StatusBadGateway, apperror.CodeServiceDown, "upstream failed", original)
	fmt.Println(err.Error())
	fmt.Println(errors.Is(err, original))
	// Output:
	// SERVICE_UNAVAILABLE: upstream failed: connection refused
	// true
}

func ExampleNotFound() {
	err := apperror.NotFound("invoice not found")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	// Output:
	// 404
	// NOT_FOUND
}

func ExampleBadRequest() {
	err := apperror.BadRequest("missing required field: email")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	// Output:
	// 400
	// BAD_REQUEST
}

func ExampleUnauthorized() {
	err := apperror.Unauthorized("invalid credentials")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	// Output:
	// 401
	// UNAUTHORIZED
}

func ExampleForbidden() {
	err := apperror.Forbidden("insufficient permissions")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	// Output:
	// 403
	// FORBIDDEN
}

func ExampleConflict() {
	err := apperror.Conflict("email already registered")
	fmt.Println(err.HTTPStatus)
	fmt.Println(err.Code)
	// Output:
	// 409
	// CONFLICT
}

func ExampleHTTPStatus() {
	err := apperror.NotFound("user not found")
	fmt.Println(apperror.HTTPStatus(err))

	fmt.Println(apperror.HTTPStatus(errors.New("generic error")))
	// Output:
	// 404
	// 500
}

func ExampleIsCode() {
	err := apperror.NotFound("user not found")
	fmt.Println(apperror.IsCode(err, apperror.CodeNotFound))
	fmt.Println(apperror.IsCode(err, apperror.CodeConflict))
	// Output:
	// true
	// false
}

func ExampleValidation() {
	err := apperror.Validation("invalid email format")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// VALIDATION_ERROR 422
}

func ExampleInternal() {
	err := apperror.Internal("unexpected failure")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// INTERNAL_ERROR 500
}

func ExampleInternalWrap() {
	cause := errors.New("db connection lost")
	err := apperror.InternalWrap("could not save", cause)
	fmt.Println(err.Code)
	fmt.Println(errors.Unwrap(err) == cause)
	// Output:
	// INTERNAL_ERROR
	// true
}

func ExampleTimeout() {
	err := apperror.Timeout("request timed out")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// TIMEOUT 504
}

func ExampleRateLimit() {
	err := apperror.RateLimit("too many requests")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// RATE_LIMIT_EXCEEDED 429
}

func ExampleServiceDown() {
	err := apperror.ServiceDown("payment gateway unreachable")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// SERVICE_UNAVAILABLE 503
}

func ExampleNotFoundWrap() {
	cause := errors.New("sql: no rows")
	err := apperror.NotFoundWrap("user not found", cause)
	fmt.Println(err.Code)
	fmt.Println(errors.Unwrap(err) == cause)
	// Output:
	// NOT_FOUND
	// true
}

func ExampleBadRequestWrap() {
	cause := errors.New("invalid JSON")
	err := apperror.BadRequestWrap("bad input", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// BAD_REQUEST 400
}

func ExampleUnauthorizedWrap() {
	cause := errors.New("token expired")
	err := apperror.UnauthorizedWrap("auth failed", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// UNAUTHORIZED 401
}

func ExampleForbiddenWrap() {
	cause := errors.New("role mismatch")
	err := apperror.ForbiddenWrap("access denied", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// FORBIDDEN 403
}

func ExampleConflictWrap() {
	cause := errors.New("duplicate key")
	err := apperror.ConflictWrap("already exists", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// CONFLICT 409
}

func ExampleValidationWrap() {
	cause := errors.New("parse error")
	err := apperror.ValidationWrap("bad input", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// VALIDATION_ERROR 422
}

func ExampleTimeoutWrap() {
	cause := errors.New("context deadline exceeded")
	err := apperror.TimeoutWrap("request timed out", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// TIMEOUT 504
}

func ExampleServiceDownWrap() {
	cause := errors.New("connection refused")
	err := apperror.ServiceDownWrap("upstream down", cause)
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// SERVICE_UNAVAILABLE 503
}

func ExamplePaymentRequired() {
	err := apperror.PaymentRequired("subscription required")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// PAYMENT_REQUIRED 402
}

func ExampleTooLarge() {
	err := apperror.TooLarge("file exceeds 10MB limit")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// PAYLOAD_TOO_LARGE 413
}

func ExampleGone() {
	err := apperror.Gone("resource has been deleted")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// GONE 410
}

func ExampleNotImplemented() {
	err := apperror.NotImplemented("feature coming soon")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// NOT_IMPLEMENTED 501
}

func ExamplePreconditionFailed() {
	err := apperror.PreconditionFailed("ETag mismatch")
	fmt.Println(err.Code, err.HTTPStatus)
	// Output:
	// PRECONDITION_FAILED 412
}

func ExampleWithStack() {
	err := apperror.NotFound("missing")
	withStack := apperror.WithStack(err)
	fmt.Println(apperror.HasStack(withStack))
	fmt.Println(apperror.HasStack(err))
	// Output:
	// true
	// false
}

func ExampleNewMultiError() {
	me := apperror.NewMultiError("validation failed",
		apperror.FieldError{Field: "email", Message: "required"},
		apperror.FieldError{Field: "name", Message: "too short"},
	)
	fmt.Println(me.HasErrors())
	fmt.Println(me.HTTPStatus)
	// Output:
	// true
	// 422
}

func ExampleMultiError_AddField() {
	me := apperror.NewMultiError("validation failed")
	me.AddField("email", "required")
	me.AddField("age", "must be positive")
	fmt.Println(len(me.Errors))
	// Output:
	// 2
}

func ExampleMultiError_OrNil() {
	me := apperror.NewMultiError("validation failed")
	fmt.Println(me.OrNil())
	me.AddField("email", "required")
	fmt.Println(me.OrNil() != nil)
	// Output:
	// <nil>
	// true
}

func ExampleMultiError_Error() {
	me := apperror.NewMultiError("validation failed",
		apperror.FieldError{Field: "email", Message: "required"},
	)
	fmt.Println(me.Error())
	// Output:
	// VALIDATION_ERROR: validation failed [email: required]
}
