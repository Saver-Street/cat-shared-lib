package cond_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/cond"
)

func ExampleTernary() {
	result := cond.Ternary(len("hello") > 3, "long", "short")
	fmt.Println(result)
	// Output: long
}

func ExampleCoalesce() {
	result := cond.Coalesce("", "", "fallback")
	fmt.Println(result)
	// Output: fallback
}

func ExampleClamp() {
	result := cond.Clamp(150, 0, 100)
	fmt.Println(result)
	// Output: 100
}

func ExampleSwitch() {
	status := 404
	msg := cond.Switch(
		cond.Case[string]{When: status == 200, Then: "OK"},
		cond.Case[string]{When: status == 404, Then: "Not Found"},
		cond.Case[string]{When: status == 500, Then: "Server Error"},
	)
	fmt.Println(msg)
	// Output: Not Found
}
