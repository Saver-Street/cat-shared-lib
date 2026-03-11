package testkit_test

import (
	"context"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func ExampleCallRecorder() {
	var cr testkit.CallRecorder

	cr.Record("hello", 42)
	cr.Record("world", 99)

	fmt.Println(cr.CallCount())
	fmt.Println(cr.Call(0))
	fmt.Println(cr.Call(1))
	// Output:
	// 2
	// [hello 42]
	// [world 99]
}

func ExampleCallRecorder_Reset() {
	var cr testkit.CallRecorder

	cr.Record("a")
	cr.Reset()

	fmt.Println(cr.CallCount())
	// Output:
	// 0
}

func ExampleMustMarshalJSON() {
	b := testkit.MustMarshalJSON(map[string]int{"x": 1})
	fmt.Println(string(b))
	// Output:
	// {"x":1}
}

func ExamplePtr() {
	s := testkit.Ptr("hello")
	fmt.Println(*s)
	n := testkit.Ptr(42)
	fmt.Println(*n)
	// Output:
	// hello
	// 42
}

func ExampleContextWithValue() {
	type ctxKey string
	ctx := testkit.ContextWithValue(context.Background(), ctxKey("user"), "alice")
	fmt.Println(ctx.Value(ctxKey("user")))
	// Output:
	// alice
}

func ExampleNewRequest() {
	r := testkit.NewRequest("GET", "/api/items?page=2", nil)
	fmt.Println(r.Method, r.URL.Path, r.URL.Query().Get("page"))
	// Output:
	// GET /api/items 2
}
