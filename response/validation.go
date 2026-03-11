package response

import (
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/validation"
)

// FieldError is a single field-level validation error in the API response.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrorResponse is the structured 422 response body.
type ValidationErrorResponse struct {
	Error  string       `json:"error"`
	Fields []FieldError `json:"fields"`
}

// ValidationErrors sends a 422 Unprocessable Entity response containing
// structured field-level errors. It accepts a slice of errors from
// [validation.Validator.Errors] or [validation.Collect].
//
// Each error that is a [*validation.ValidationError] is included with its
// field name and message. Other error types are included with an empty field
// and their Error() text.
func ValidationErrors(w http.ResponseWriter, errs []error) {
	fields := make([]FieldError, 0, len(errs))
	for _, err := range errs {
		if err == nil {
			continue
		}
		if ve, ok := err.(*validation.ValidationError); ok {
			fields = append(fields, FieldError{
				Field:   ve.Field,
				Message: ve.Message,
			})
		} else {
			fields = append(fields, FieldError{
				Message: err.Error(),
			})
		}
	}
	JSON(w, http.StatusUnprocessableEntity, ValidationErrorResponse{
		Error:  "Validation failed",
		Fields: fields,
	})
}
