package request_test

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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

func ExampleParseCursor() {
	q := url.Values{"cursor": {"abc123"}, "limit": {"25"}}
	cp := request.ParseCursor(q, 20, 100)
	fmt.Printf("Cursor=%s Limit=%d\n", cp.Cursor, cp.Limit)
	// Output:
	// Cursor=abc123 Limit=25
}

func ExampleParseCursor_defaults() {
	q := url.Values{}
	cp := request.ParseCursor(q, 20, 100)
	fmt.Printf("Cursor=%q Limit=%d\n", cp.Cursor, cp.Limit)
	// Output:
	// Cursor="" Limit=20
}

func ExampleExtractBearerToken() {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer my-token-123")
	token, ok := request.ExtractBearerToken(r)
	fmt.Println(token, ok)
	// Output:
	// my-token-123 true
}

func ExampleContentType() {
	r, _ := http.NewRequest("POST", "/", strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json; charset=utf-8")
	fmt.Println(request.ContentType(r))
	// Output:
	// application/json
}

func ExampleIsJSON() {
	r, _ := http.NewRequest("POST", "/", strings.NewReader("{}"))
	r.Header.Set("Content-Type", "application/json")
	fmt.Println(request.IsJSON(r))
	// Output:
	// true
}

func ExampleClientIP() {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.50, 70.41.3.18")
	fmt.Println(request.ClientIP(r))
	// Output:
	// 203.0.113.50
}

func ExampleOptionalQueryFloat() {
	q := url.Values{"rating": {"4.5"}}
	fmt.Println(request.OptionalQueryFloat(q, "rating", 0.0))
	fmt.Println(request.OptionalQueryFloat(q, "score", 1.0))
	// Output:
	// 4.5
	// 1
}

func ExampleOptionalQueryBool() {
	q := url.Values{"active": {"true"}}
	fmt.Println(*request.OptionalQueryBool(q, "active"))
	fmt.Println(request.OptionalQueryBool(q, "deleted"))
	// Output:
	// true
	// <nil>
}

func ExampleRequireHeader() {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("X-Tenant-ID", "acme")
	val, err := request.RequireHeader(r, "X-Tenant-ID")
	fmt.Println(val, err)
	// Output:
	// acme <nil>
}

func ExampleParseIDList() {
	q := url.Values{"ids": {"550e8400-e29b-41d4-a716-446655440000,6ba7b810-9dad-11d1-80b4-00c04fd430c8"}}
	ids, err := request.ParseIDList(q, "ids")
	fmt.Println(len(ids), err)
	// Output:
	// 2 <nil>
}
