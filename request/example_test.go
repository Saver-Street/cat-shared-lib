package request_test

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/Saver-Street/cat-shared-lib/request"
)

func ExampleParsePagination() {
	q := url.Values{"page": {"3"}, "limit": {"25"}}
	p := request.ParsePagination(q, 20, 100)
	fmt.Printf("Page=%d Limit=%d Offset=%d\n", p.Page, p.Limit, p.Offset)
	// Output:
	// Page=3 Limit=25 Offset=50
}

func ExampleParsePagination_defaults() {
	p := request.ParsePagination(url.Values{}, 20, 100)
	fmt.Printf("Page=%d Limit=%d Offset=%d\n", p.Page, p.Limit, p.Offset)
	// Output:
	// Page=1 Limit=20 Offset=0
}

func ExampleRequireURLParam() {
	paramFn := func(r *http.Request, key string) string {
		return r.URL.Query().Get(key)
	}
	r, _ := http.NewRequest("GET", "/test?id=abc-123", nil)
	val, err := request.RequireURLParam(r, "id", paramFn)
	fmt.Println(val, err)
	// Output:
	// abc-123 <nil>
}

func ExampleRequireURLParamInt() {
	paramFn := func(r *http.Request, key string) string {
		return r.URL.Query().Get(key)
	}
	r, _ := http.NewRequest("GET", "/test?id=42", nil)
	n, err := request.RequireURLParamInt(r, "id", paramFn)
	fmt.Println(n, err)
	// Output:
	// 42 <nil>
}

func ExampleRequireQueryParam() {
	q := url.Values{"status": {"active"}}
	val, err := request.RequireQueryParam(q, "status")
	fmt.Println(val, err)
	// Output:
	// active <nil>
}

func ExampleParseBoolParam() {
	q := url.Values{"debug": {"true"}}
	fmt.Println(request.ParseBoolParam(q, "debug", false))
	fmt.Println(request.ParseBoolParam(url.Values{}, "debug", false))
	// Output:
	// true
	// false
}

func ExampleOptionalQueryParam() {
	q := url.Values{"status": {"active"}}
	fmt.Println(request.OptionalQueryParam(q, "status", "pending"))
	fmt.Println(request.OptionalQueryParam(q, "missing", "pending"))
	// Output:
	// active
	// pending
}

func ExampleOptionalQueryInt() {
	q := url.Values{"limit": {"50"}}
	fmt.Println(request.OptionalQueryInt(q, "limit", 25))
	fmt.Println(request.OptionalQueryInt(q, "missing", 25))
	// Output:
	// 50
	// 25
}

func ExampleRequireQueryParamInt() {
	q := url.Values{"age": {"30"}}
	val, err := request.RequireQueryParamInt(q, "age")
	fmt.Println(val, err)
	// Output:
	// 30 <nil>
}

func ExampleParseCommaSeparated() {
	q := url.Values{"tags": {"go,rust,python"}}
	fmt.Println(request.ParseCommaSeparated(q, "tags"))
	fmt.Println(request.ParseCommaSeparated(q, "missing"))
	// Output:
	// [go rust python]
	// []
}

func ExampleParseSortOrder() {
	q := url.Values{"sort": {"name"}, "order": {"desc"}}
	field, dir := request.ParseSortOrder(q, []string{"name", "created_at"}, "created_at", "asc")
	fmt.Println(field, dir)
	// Output:
	// name desc
}

func ExampleParseEnumParam() {
	q := url.Values{"role": {"admin"}}
	fmt.Println(request.ParseEnumParam(q, "role", []string{"admin", "user", "guest"}, "user"))
	fmt.Println(request.ParseEnumParam(q, "missing", []string{"admin", "user"}, "user"))
	// Output:
	// admin
	// user
}

func ExampleParseCommaSeparatedInts() {
	q := url.Values{"ids": {"1,2,3,4"}}
	ids, err := request.ParseCommaSeparatedInts(q, "ids")
	fmt.Println(ids, err)
	// Output:
	// [1 2 3 4] <nil>
}

func ExampleParseDateParam() {
	q := url.Values{"since": {"2024-01-15"}}
	t, err := request.ParseDateParam(q, "since")
	fmt.Println(t.Format("2006-01-02"), err)
	// Output:
	// 2024-01-15 <nil>
}

func ExampleRequireUUIDParam() {
	paramFn := func(r *http.Request, key string) string {
		return "550e8400-e29b-41d4-a716-446655440000"
	}
	r, _ := http.NewRequest("GET", "/users/550e8400-e29b-41d4-a716-446655440000", nil)
	id, err := request.RequireUUIDParam(r, "id", paramFn)
	fmt.Println(id, err)
	// Output:
	// 550e8400-e29b-41d4-a716-446655440000 <nil>
}
