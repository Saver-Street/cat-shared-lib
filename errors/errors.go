// Package errors provides standardized error types with HTTP status code
// mapping for use across microservices.
package errors

import (
	"fmt"
	"net/http"
)

// Code represents a standardized application error code.
type Code string

const (
	CodeBadRequest          Code = "BAD_REQUEST"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeNotFound            Code = "NOT_FOUND"
	CodeConflict            Code = "CONFLICT"
	CodeUnprocessableEntity Code = "UNPROCESSABLE_ENTITY"
	CodeTooManyRequests     Code = "TOO_MANY_REQUESTS"
	CodeInternal            Code = "INTERNAL_ERROR"
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"
	CodeGatewayTimeout      Code = "GATEWAY_TIMEOUT"
	CodeValidation          Code = "VALIDATION_ERROR"
)

// httpStatus maps error codes to HTTP status codes.
var httpStatus = map[Code]int{
	CodeBadRequest:          http.StatusBadRequest,
	CodeUnauthorized:        http.StatusUnauthorized,
	CodeForbidden:           http.StatusForbidden,
	CodeNotFound:            http.StatusNotFound,
	CodeConflict:            http.StatusConflict,
	CodeUnprocessableEntity: http.StatusUnprocessableEntity,
	CodeTooManyRequests:     http.StatusTooManyRequests,
	CodeInternal:            http.StatusInternalServerError,
	CodeServiceUnavailable:  http.StatusServiceUnavailable,
	CodeGatewayTimeout:      http.StatusGatewayTimeout,
	CodeValidation:          http.StatusUnprocessableEntity,
}

// AppError is a structured application error that carries an error code,
// a human-readable message, and an optional underlying error.
type AppError struct {
	// Code is the standardized error code.
	Code Code `json:"code"`
	// Message is the user-facing error message.
	Message string `json:"message"`
	// Err is the underlying error, if any. Not serialized.
	Err error `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is / errors.As support.
func (e *AppError) Unwrap() error {
	return e.Err
}

// HTTPStatus returns the HTTP status code for this error.
func (e *AppError) HTTPStatus() int {
	if s, ok := httpStatus[e.Code]; ok {
		return s
	}
	return http.StatusInternalServerError
}

// New creates a new AppError with the given code and message.
func New(code Code, message string) *AppError {
	return &AppError{Code: code, Message: message}
}

// Wrap creates a new AppError wrapping an existing error.
func Wrap(code Code, message string, err error) *AppError {
	return &AppError{Code: code, Message: message, Err: err}
}

// HTTPStatusForCode returns the HTTP status code for the given error code.
// Returns 500 for unknown codes.
func HTTPStatusForCode(code Code) int {
	if s, ok := httpStatus[code]; ok {
		return s
	}
	return http.StatusInternalServerError
}

// convenience constructors

// BadRequest creates a BAD_REQUEST error.
func BadRequest(msg string) *AppError { return New(CodeBadRequest, msg) }

// Unauthorized creates an UNAUTHORIZED error.
func Unauthorized(msg string) *AppError { return New(CodeUnauthorized, msg) }

// Forbidden creates a FORBIDDEN error.
func Forbidden(msg string) *AppError { return New(CodeForbidden, msg) }

// NotFound creates a NOT_FOUND error.
func NotFound(msg string) *AppError { return New(CodeNotFound, msg) }

// Conflict creates a CONFLICT error.
func Conflict(msg string) *AppError { return New(CodeConflict, msg) }

// Internal creates an INTERNAL_ERROR error.
func Internal(msg string) *AppError { return New(CodeInternal, msg) }

// Validation creates a VALIDATION_ERROR error.
func Validation(msg string) *AppError { return New(CodeValidation, msg) }
