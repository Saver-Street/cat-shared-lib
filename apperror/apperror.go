// Package apperror provides standardized application error types with HTTP
// status codes and machine-readable error codes. Use these to return
// consistent error responses across microservices.
package apperror

import (
	"errors"
	"fmt"
	"net/http"
)

// Compile-time interface compliance check.
var _ error = (*Error)(nil)

// Code is a machine-readable error code string.
type Code string

// Standard error codes used across services.
const (
	CodeNotFound         Code = "NOT_FOUND"
	CodeBadRequest       Code = "BAD_REQUEST"
	CodeUnauthorized     Code = "UNAUTHORIZED"
	CodeForbidden        Code = "FORBIDDEN"
	CodeConflict         Code = "CONFLICT"
	CodeValidation       Code = "VALIDATION_ERROR"
	CodeInternal         Code = "INTERNAL_ERROR"
	CodeTimeout          Code = "TIMEOUT"
	CodeRateLimit        Code = "RATE_LIMIT_EXCEEDED"
	CodeServiceDown      Code = "SERVICE_UNAVAILABLE"
	CodeUnprocessable    Code = "UNPROCESSABLE_ENTITY"
	CodeMethodNotAllowed Code = "METHOD_NOT_ALLOWED"
	CodePaymentRequired  Code = "PAYMENT_REQUIRED"
	CodeTooLarge         Code = "PAYLOAD_TOO_LARGE"
	CodeGone             Code = "GONE"
	CodeNotImplemented   Code = "NOT_IMPLEMENTED"
	CodePrecondition     Code = "PRECONDITION_FAILED"
)

// Error is a structured application error with an HTTP status code,
// a machine-readable code, and a human-readable message.
type Error struct {
	// HTTPStatus is the HTTP status code to return (e.g. 404).
	HTTPStatus int `json:"-"`
	// Code is a machine-readable identifier (e.g. "NOT_FOUND").
	Code Code `json:"code"`
	// Message is a human-readable description of the error.
	Message string `json:"message"`
	// Err is the underlying error, if any. Not serialized to JSON.
	Err error `json:"-"`
	// Stack is the captured stack trace, if any. Not serialized to JSON.
	Stack StackTrace `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/errors.As support.
func (e *Error) Unwrap() error {
	return e.Err
}

// New creates a new Error with the given status, code, and message.
func New(status int, code Code, message string) *Error {
	return &Error{HTTPStatus: status, Code: code, Message: message}
}

// Wrap creates a new Error wrapping an underlying error.
func Wrap(status int, code Code, message string, err error) *Error {
	return &Error{HTTPStatus: status, Code: code, Message: message, Err: err}
}

// NotFound creates a 404 NOT_FOUND error.
func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message)
}

// BadRequest creates a 400 BAD_REQUEST error.
func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized creates a 401 UNAUTHORIZED error.
func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden creates a 403 FORBIDDEN error.
func Forbidden(message string) *Error {
	return New(http.StatusForbidden, CodeForbidden, message)
}

// Conflict creates a 409 CONFLICT error.
func Conflict(message string) *Error {
	return New(http.StatusConflict, CodeConflict, message)
}

// Validation creates a 422 VALIDATION_ERROR error.
func Validation(message string) *Error {
	return New(http.StatusUnprocessableEntity, CodeValidation, message)
}

// Internal creates a 500 INTERNAL_ERROR error.
func Internal(message string) *Error {
	return New(http.StatusInternalServerError, CodeInternal, message)
}

// InternalWrap creates a 500 INTERNAL_ERROR wrapping an underlying error.
func InternalWrap(message string, err error) *Error {
	return Wrap(http.StatusInternalServerError, CodeInternal, message, err)
}

// Timeout creates a 504 TIMEOUT error.
func Timeout(message string) *Error {
	return New(http.StatusGatewayTimeout, CodeTimeout, message)
}

// RateLimit creates a 429 RATE_LIMIT_EXCEEDED error.
func RateLimit(message string) *Error {
	return New(http.StatusTooManyRequests, CodeRateLimit, message)
}

// ServiceDown creates a 503 SERVICE_UNAVAILABLE error.
func ServiceDown(message string) *Error {
	return New(http.StatusServiceUnavailable, CodeServiceDown, message)
}

// NotFoundWrap creates a 404 NOT_FOUND error wrapping an underlying error.
func NotFoundWrap(message string, err error) *Error {
	return Wrap(http.StatusNotFound, CodeNotFound, message, err)
}

// BadRequestWrap creates a 400 BAD_REQUEST error wrapping an underlying error.
func BadRequestWrap(message string, err error) *Error {
	return Wrap(http.StatusBadRequest, CodeBadRequest, message, err)
}

// UnauthorizedWrap creates a 401 UNAUTHORIZED error wrapping an underlying error.
func UnauthorizedWrap(message string, err error) *Error {
	return Wrap(http.StatusUnauthorized, CodeUnauthorized, message, err)
}

// ForbiddenWrap creates a 403 FORBIDDEN error wrapping an underlying error.
func ForbiddenWrap(message string, err error) *Error {
	return Wrap(http.StatusForbidden, CodeForbidden, message, err)
}

// ConflictWrap creates a 409 CONFLICT error wrapping an underlying error.
func ConflictWrap(message string, err error) *Error {
	return Wrap(http.StatusConflict, CodeConflict, message, err)
}

// ValidationWrap creates a 422 VALIDATION_ERROR error wrapping an underlying error.
func ValidationWrap(message string, err error) *Error {
	return Wrap(http.StatusUnprocessableEntity, CodeValidation, message, err)
}

// TimeoutWrap creates a 504 TIMEOUT error wrapping an underlying error.
func TimeoutWrap(message string, err error) *Error {
	return Wrap(http.StatusGatewayTimeout, CodeTimeout, message, err)
}

// ServiceDownWrap creates a 503 SERVICE_UNAVAILABLE error wrapping an underlying error.
func ServiceDownWrap(message string, err error) *Error {
	return Wrap(http.StatusServiceUnavailable, CodeServiceDown, message, err)
}

// PaymentRequired creates a 402 PAYMENT_REQUIRED error.
func PaymentRequired(message string) *Error {
	return New(http.StatusPaymentRequired, CodePaymentRequired, message)
}

// TooLarge creates a 413 PAYLOAD_TOO_LARGE error.
func TooLarge(message string) *Error {
	return New(http.StatusRequestEntityTooLarge, CodeTooLarge, message)
}

// Gone creates a 410 GONE error.
func Gone(message string) *Error {
	return New(http.StatusGone, CodeGone, message)
}

// NotImplemented creates a 501 NOT_IMPLEMENTED error.
func NotImplemented(message string) *Error {
	return New(http.StatusNotImplemented, CodeNotImplemented, message)
}

// PreconditionFailed creates a 412 PRECONDITION_FAILED error.
func PreconditionFailed(message string) *Error {
	return New(http.StatusPreconditionFailed, CodePrecondition, message)
}

// HTTPStatus returns the HTTP status code for an error. If the error is
// an *Error it returns its HTTPStatus field; otherwise it returns 500.
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	var e *Error
	if errors.As(err, &e) {
		return e.HTTPStatus
	}
	return http.StatusInternalServerError
}

// IsCode checks whether err is an *Error with the given code.
func IsCode(err error, code Code) bool {

	var e *Error
	if errors.As(err, &e) {
		return e.Code == code
	}
	return false
}
