package apperror

import (
	"fmt"
	"net/http"
	"strings"
)

// FieldError represents a validation error on a single field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// MultiError collects multiple field-level validation errors into a single
// error that can be returned as a structured JSON response.
type MultiError struct {
	HTTPStatus int          `json:"-"`
	Code       Code         `json:"code"`
	Message    string       `json:"message"`
	Errors     []FieldError `json:"errors"`
}

// Error implements the error interface.
func (me *MultiError) Error() string {
	msgs := make([]string, len(me.Errors))
	for i, fe := range me.Errors {
		msgs[i] = fmt.Sprintf("%s: %s", fe.Field, fe.Message)
	}
	return fmt.Sprintf("%s: %s [%s]", me.Code, me.Message, strings.Join(msgs, "; "))
}

// NewMultiError creates a MultiError with the VALIDATION_ERROR code and 422 status.
func NewMultiError(message string, errs ...FieldError) *MultiError {
	return &MultiError{
		HTTPStatus: http.StatusUnprocessableEntity,
		Code:       CodeValidation,
		Message:    message,
		Errors:     errs,
	}
}

// AddField appends a field error.
func (me *MultiError) AddField(field, message string) {
	me.Errors = append(me.Errors, FieldError{Field: field, Message: message})
}

// HasErrors reports whether the MultiError contains any field errors.
func (me *MultiError) HasErrors() bool {
	return len(me.Errors) > 0
}

// OrNil returns nil if no field errors exist, otherwise returns the MultiError.
// This is convenient for validation functions that accumulate errors.
func (me *MultiError) OrNil() error {
	if len(me.Errors) == 0 {
		return nil
	}
	return me
}
