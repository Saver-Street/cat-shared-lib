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
