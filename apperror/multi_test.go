package apperror

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewMultiError(t *testing.T) {
	fe := FieldError{Field: "email", Message: "is required"}
	me := NewMultiError("validation failed", fe)

	testkit.AssertEqual(t, me.HTTPStatus, http.StatusUnprocessableEntity)
	testkit.AssertEqual(t, me.Code, CodeValidation)
	testkit.AssertEqual(t, me.Message, "validation failed")
	testkit.AssertLen(t, me.Errors, 1)
	testkit.AssertEqual(t, me.Errors[0].Field, "email")
	testkit.AssertEqual(t, me.Errors[0].Message, "is required")
}

func TestNewMultiError_NoFields(t *testing.T) {
	me := NewMultiError("validation failed")

	testkit.AssertEqual(t, me.HTTPStatus, http.StatusUnprocessableEntity)
	testkit.AssertEqual(t, me.Code, CodeValidation)
	testkit.AssertLen(t, me.Errors, 0)
}

func TestMultiError_AddField(t *testing.T) {
	me := NewMultiError("validation failed")
	me.AddField("name", "too short")
	me.AddField("age", "must be positive")

	testkit.AssertLen(t, me.Errors, 2)
	testkit.AssertEqual(t, me.Errors[0].Field, "name")
	testkit.AssertEqual(t, me.Errors[0].Message, "too short")
	testkit.AssertEqual(t, me.Errors[1].Field, "age")
	testkit.AssertEqual(t, me.Errors[1].Message, "must be positive")
}

func TestMultiError_HasErrors(t *testing.T) {
	me := NewMultiError("validation failed")
	testkit.AssertFalse(t, me.HasErrors())

	me.AddField("email", "invalid")
	testkit.AssertTrue(t, me.HasErrors())
}

func TestMultiError_OrNil_Empty(t *testing.T) {
	me := NewMultiError("validation failed")
	testkit.AssertNil(t, me.OrNil())
}

func TestMultiError_OrNil_WithErrors(t *testing.T) {
	me := NewMultiError("validation failed")
	me.AddField("email", "is required")

	err := me.OrNil()
	testkit.AssertNotNil(t, err)
	testkit.AssertEqual(t, err, me)
}

func TestMultiError_Error(t *testing.T) {
	me := NewMultiError("validation failed",
		FieldError{Field: "email", Message: "is required"},
		FieldError{Field: "name", Message: "too short"},
	)

	got := me.Error()
	testkit.AssertContains(t, got, "VALIDATION_ERROR")
	testkit.AssertContains(t, got, "validation failed")
	testkit.AssertContains(t, got, "email: is required")
	testkit.AssertContains(t, got, "name: too short")
	testkit.AssertEqual(t, got, "VALIDATION_ERROR: validation failed [email: is required; name: too short]")
}

func TestMultiError_Error_Single(t *testing.T) {
	me := NewMultiError("bad input", FieldError{Field: "id", Message: "missing"})
	testkit.AssertEqual(t, me.Error(), "VALIDATION_ERROR: bad input [id: missing]")
}

func TestMultiError_JSONSerialization(t *testing.T) {
	me := NewMultiError("validation failed",
		FieldError{Field: "email", Message: "is required"},
		FieldError{Field: "name", Message: "too short"},
	)

	data, err := json.Marshal(me)
	testkit.AssertNoError(t, err)

	var got map[string]any
	testkit.AssertNoError(t, json.Unmarshal(data, &got))

	testkit.AssertEqual(t, got["code"], string(CodeValidation))
	testkit.AssertEqual(t, got["message"], "validation failed")

	// HTTPStatus should not be in JSON (json:"-")
	_, hasStatus := got["HTTPStatus"]
	testkit.AssertFalse(t, hasStatus)

	errs, ok := got["errors"].([]any)
	testkit.AssertTrue(t, ok)
	testkit.AssertLen(t, errs, 2)

	first := errs[0].(map[string]any)
	testkit.AssertEqual(t, first["field"], "email")
	testkit.AssertEqual(t, first["message"], "is required")
}

func TestMultiError_Empty(t *testing.T) {
	me := NewMultiError("no errors")
	testkit.AssertFalse(t, me.HasErrors())
	testkit.AssertNil(t, me.OrNil())
	testkit.AssertEqual(t, me.Error(), "VALIDATION_ERROR: no errors []")
}
