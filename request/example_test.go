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
