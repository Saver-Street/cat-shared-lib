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
