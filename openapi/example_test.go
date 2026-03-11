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
