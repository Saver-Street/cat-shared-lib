package sanitize_test

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
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
	pgDup := &pgconn.PgError{Code: "23505"}
	fmt.Println(sanitize.IsDuplicateKey(pgDup))
	_ = errors.New("unused") // keep errors import
	fmt.Println(sanitize.IsDuplicateKey(errors.New("other error")))
	fmt.Println(sanitize.IsDuplicateKey(nil))
	// Output:
	// true
	// false
	// false
}
