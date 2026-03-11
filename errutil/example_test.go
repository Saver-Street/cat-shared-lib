package errutil_test

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Saver-Street/cat-shared-lib/errutil"
)

func ExampleMust() {
	v := errutil.Must(strconv.Atoi("42"))
	fmt.Println(v)
	// Output: 42
}

func ExampleCombine() {
	err := errutil.Combine(
		nil,
		errors.New("first"),
		nil,
		errors.New("second"),
	)
	fmt.Println(err)
	// Output:
	// first
	// second
}

func ExampleWrap() {
	base := errors.New("file not found")
	err := errutil.Wrap(base, "open config")
	fmt.Println(err)
	// Output: open config: file not found
}

func ExampleRecover() {
	err := errutil.Recover(func() {
		panic("something went wrong")
	})
	fmt.Println(err)
	// Output: recovered panic: something went wrong
}
