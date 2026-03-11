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

func ExampleNewSet() {
	s := types.NewSet("go", "rust", "python", "go")
	fmt.Println(s.Len())
	fmt.Println(s.Contains("go"))
	fmt.Println(s.Contains("java"))
	// Output:
	// 3
	// true
	// false
}

func ExampleSet_Union() {
	a := types.NewSet(1, 2, 3)
	b := types.NewSet(3, 4, 5)
	u := a.Union(b)
	fmt.Println(u.Len())
	// Output:
	// 5
}
