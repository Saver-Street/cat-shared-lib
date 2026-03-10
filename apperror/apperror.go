// Package apperror provides standardized application error types with HTTP
// status codes and machine-readable error codes. Use these to return
// consistent error responses across microservices.
package apperror

import (
	"fmt"
	"net/http"
)

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

// HTTPStatus returns the HTTP status code for an error. If the error is
// an *Error it returns its HTTPStatus field; otherwise it returns 500.
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if e, ok := err.(*Error); ok {
		return e.HTTPStatus
	}
	return http.StatusInternalServerError
}

// IsCode checks whether err is an *Error with the given code.
func IsCode(err error, code Code) bool {
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return false
}
