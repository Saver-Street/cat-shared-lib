package sanitize_test

import (
	"errors"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/sanitize"
)

func ExampleDocFilename() {
	fmt.Println(sanitize.DocFilename("report.pdf"))
	fmt.Println(sanitize.DocFilename("../../../etc/passwd"))
	fmt.Println(sanitize.DocFilename(""))
	// Output:
	// report.pdf
	// passwd
	// unnamed
}

func ExampleNilIfEmpty() {
	fmt.Println(sanitize.NilIfEmpty(""))
	fmt.Println(*sanitize.NilIfEmpty("hello"))
	// Output:
	// <nil>
	// hello
}

func ExampleIsDuplicateKey() {
	err := errors.New("ERROR: duplicate key value violates unique constraint (SQLSTATE 23505)")
	fmt.Println(sanitize.IsDuplicateKey(err))
	fmt.Println(sanitize.IsDuplicateKey(errors.New("other error")))
	fmt.Println(sanitize.IsDuplicateKey(nil))
	// Output:
	// true
	// false
	// false
}
