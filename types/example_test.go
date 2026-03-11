package types_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/types"
)

func ExampleNormalizePage() {
	p := types.NormalizePage(3, 25)
	fmt.Printf("Page=%d Limit=%d Offset=%d\n", p.Page, p.Limit, p.Offset)
	// Output:
	// Page=3 Limit=25 Offset=50
}

func ExampleNormalizePage_defaults() {
	p := types.NormalizePage(0, 0)
	fmt.Printf("Page=%d Limit=%d Offset=%d\n", p.Page, p.Limit, p.Offset)
	// Output:
	// Page=1 Limit=20 Offset=0
}

func ExampleNormalizePage_capLimit() {
	p := types.NormalizePage(1, 500)
	fmt.Printf("Limit=%d\n", p.Limit)
	// Output:
	// Limit=100
}

func ExamplePaginationParams_HasNextPage() {
	p := types.NormalizePage(1, 10)
	fmt.Println(p.HasNextPage(25))
	fmt.Println(p.HasNextPage(10))
	// Output:
	// true
	// false
}

func ExamplePaginationParams_IsLastPage() {
	p := types.NormalizePage(2, 10)
	fmt.Println(p.IsLastPage(20))
	fmt.Println(p.IsLastPage(25))
	// Output:
	// true
	// false
}

func ExamplePaginationParams_TotalPages() {
p := types.NormalizePage(1, 10)
fmt.Println(p.TotalPages(25))
fmt.Println(p.TotalPages(0))
// Output:
// 3
// 0
}

func ExampleNormalizeCursor() {
cp := types.NormalizeCursor("abc123", 50)
fmt.Printf("Cursor=%s Limit=%d\n", cp.Cursor, cp.Limit)
// Output:
// Cursor=abc123 Limit=50
}

func ExampleNormalizeCursor_defaults() {
cp := types.NormalizeCursor("", 0)
fmt.Printf("Cursor=%q Limit=%d\n", cp.Cursor, cp.Limit)
// Output:
// Cursor="" Limit=20
}

func ExampleNewCursorPage() {
items := []string{"a", "b", "c", "d", "e"}
page := types.NewCursorPage(items, 3, func(s string) string { return s })
fmt.Printf("Items=%v NextCursor=%s HasMore=%v\n", page.Items, page.NextCursor, page.HasMore)
// Output:
// Items=[a b c] NextCursor=c HasMore=true
}

func ExampleNewCursorPage_lastPage() {
items := []string{"a", "b"}
page := types.NewCursorPage(items, 5, func(s string) string { return s })
fmt.Printf("Items=%v NextCursor=%q HasMore=%v\n", page.Items, page.NextCursor, page.HasMore)
// Output:
// Items=[a b] NextCursor="" HasMore=false
}

func ExampleApplyOffset() {
items := []int{10, 20, 30, 40, 50}
fmt.Println(types.ApplyOffset(items, 1, 2))
fmt.Println(types.ApplyOffset(items, 10, 2))
// Output:
// [20 30]
// []
}
