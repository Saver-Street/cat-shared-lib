package sanitize_test

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

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
	fmt.Println(sanitize.IsDuplicateKey(errors.New("other error")))
	fmt.Println(sanitize.IsDuplicateKey(nil))
	// Output:
	// true
	// false
	// false
}

func ExampleTruncateFilename() {
	fmt.Println(sanitize.TruncateFilename("report.pdf", 20))
	fmt.Println(sanitize.TruncateFilename("averylongfilename.txt", 10))
	fmt.Println(sanitize.TruncateFilename("", 10))
	fmt.Println(sanitize.TruncateFilename("file.txt", 0))
	// Output:
	// report.pdf
	// averyl.txt
	//
	//
}

func ExampleMaxLength() {
	fmt.Println(sanitize.MaxLength("hello world", 5))
	fmt.Println(sanitize.MaxLength("hi", 10))
	fmt.Println(sanitize.MaxLength("hello", 0))
	// Output:
	// hello
	// hi
	//
}

func ExampleSanitizeEmail() {
	fmt.Println(sanitize.SanitizeEmail("  User@Example.COM  "))
	fmt.Println(sanitize.SanitizeEmail("alice@example.com"))
	fmt.Println(sanitize.SanitizeEmail(""))
	// Output:
	// user@example.com
	// alice@example.com
	//
}

func ExampleIsDatabaseError() {
	pgErr := &pgconn.PgError{Code: "23503"}
	fmt.Println(sanitize.IsDatabaseError(pgErr, "23503"))
	fmt.Println(sanitize.IsDatabaseError(pgErr, "23505"))
	fmt.Println(sanitize.IsDatabaseError(nil, "23505"))
	// Output:
	// true
	// false
	// false
}

func ExampleTrimAndNilIfEmpty() {
	fmt.Println(sanitize.TrimAndNilIfEmpty(""))
	fmt.Println(sanitize.TrimAndNilIfEmpty("   "))
	fmt.Println(*sanitize.TrimAndNilIfEmpty("  hello  "))
	// Output:
	// <nil>
	// <nil>
	// hello
}
