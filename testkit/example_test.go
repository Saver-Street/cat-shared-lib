package testkit_test

import (
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
