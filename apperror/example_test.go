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
