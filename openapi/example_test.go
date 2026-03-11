package openapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/openapi"
)

func ExampleNewSpec() {
	spec := openapi.NewSpec("Billing API", "1.0.0").
		AddPath("/invoices", http.MethodGet, &openapi.Operation{
			Summary:     "List invoices",
			OperationID: "listInvoices",
			Tags:        []string{"invoices"},
		})

	rec := httptest.NewRecorder()
	spec.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))

	fmt.Println(rec.Code)
	fmt.Println(strings.Contains(rec.Body.String(), "Billing API"))
	// Output:
	// 200
	// true
}

func ExampleSpec_AddSchema() {
	spec := openapi.NewSpec("Users API", "1.0.0").
		AddSchema("User", openapi.ObjectSchema(map[string]*openapi.Schema{
			"id":   openapi.IntegerSchema(),
			"name": openapi.StringSchema(),
		})).
		AddPath("/users/{id}", "get", openapi.NewOperation("Get user").
			AddResponse("200", "Success", openapi.RefSchema("#/components/schemas/User")))

	data, _ := spec.JSON()
	fmt.Println(strings.Contains(string(data), "components"))
	fmt.Println(strings.Contains(string(data), "#/components/schemas/User"))
	// Output:
	// true
	// true
}

func ExampleSpec_AddSecurityScheme() {
	spec := openapi.NewSpec("Secure API", "1.0.0").
		AddSecurityScheme("bearerAuth", openapi.BearerAuth("JWT")).
		AddPath("/protected", "get", openapi.NewOperation("Protected").
			WithSecurity("bearerAuth").
			AddResponse("200", "OK", nil))

	data, _ := spec.JSON()
	fmt.Println(strings.Contains(string(data), "securitySchemes"))
	fmt.Println(strings.Contains(string(data), "bearer"))
	// Output:
	// true
	// true
}
