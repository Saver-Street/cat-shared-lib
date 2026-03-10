package response_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/response"
)

func ExampleJSON() {
	w := httptest.NewRecorder()
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 200
	// {"status":"ok"}
}

func ExampleOK() {
	w := httptest.NewRecorder()
	response.OK(w, map[string]int{"count": 42})
	fmt.Println(w.Code)
	// Output:
	// 200
}

func ExampleCreated() {
	w := httptest.NewRecorder()
	response.Created(w, map[string]string{"id": "new-123"})
	fmt.Println(w.Code)
	// Output:
	// 201
}

func ExampleError() {
	w := httptest.NewRecorder()
	response.Error(w, http.StatusNotFound, "resource not found")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 404
	// {"error":"resource not found"}
}

func ExampleDecodeJSON() {
	body := `{"name":"Alice","age":30}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	var data map[string]any
	err := response.DecodeJSON(r, &data)
	fmt.Println(err)
	fmt.Println(data["name"])
	// Output:
	// <nil>
	// Alice
}

func ExampleDecodeOrFail() {
	body := `{"valid":"json"}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	w := httptest.NewRecorder()
	var data map[string]string
	ok := response.DecodeOrFail(w, r, &data)
	fmt.Println(ok)
	fmt.Println(data["valid"])
	// Output:
	// true
	// json
}

func ExampleAccepted() {
	w := httptest.NewRecorder()
	response.Accepted(w, map[string]string{"status": "queued"})
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 202
	// {"status":"queued"}
}

func ExampleNoContent() {
	w := httptest.NewRecorder()
	response.NoContent(w)
	fmt.Println(w.Code)
	// Output:
	// 204
}

func ExampleBadRequest() {
	w := httptest.NewRecorder()
	response.BadRequest(w, "invalid input")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 400
	// {"error":"invalid input"}
}

func ExampleUnauthorized() {
	w := httptest.NewRecorder()
	response.Unauthorized(w, "token expired")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 401
	// {"error":"token expired"}
}

func ExampleForbidden() {
	w := httptest.NewRecorder()
	response.Forbidden(w, "access denied")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 403
	// {"error":"access denied"}
}

func ExampleNotFound() {
	w := httptest.NewRecorder()
	response.NotFound(w, "item not found")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 404
	// {"error":"item not found"}
}

func ExampleConflict() {
	w := httptest.NewRecorder()
	response.Conflict(w, "already exists")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 409
	// {"error":"already exists"}
}

func ExampleUnprocessableEntity() {
	w := httptest.NewRecorder()
	response.UnprocessableEntity(w, "validation failed")
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 422
	// {"error":"validation failed"}
}

func ExampleTooManyRequests() {
	w := httptest.NewRecorder()
	response.TooManyRequests(w, "slow down")
	fmt.Println(w.Code)
	// Output:
	// 429
}

func ExampleInternalError() {
	w := httptest.NewRecorder()
	response.InternalError(w, "db connection lost", nil)
	fmt.Println(w.Code)
	fmt.Println(strings.TrimSpace(w.Body.String()))
	// Output:
	// 500
	// {"error":"Internal server error"}
}

func ExampleServiceUnavailable() {
	w := httptest.NewRecorder()
	response.ServiceUnavailable(w, "down for maintenance")
	fmt.Println(w.Code)
	// Output:
	// 503
}
