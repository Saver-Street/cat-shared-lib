package stringutil_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/stringutil"
)

func ExampleToKebabCase() {
	fmt.Println(stringutil.ToKebabCase("MyServiceName"))
	// Output: my-service-name
}

func ExamplePadLeft() {
	fmt.Println(stringutil.PadLeft("42", 5, '0'))
	// Output: 00042
}

func ExampleReverse() {
	fmt.Println(stringutil.Reverse("hello"))
	// Output: olleh
}

func ExampleWordWrap() {
	fmt.Println(stringutil.WordWrap("the quick brown fox jumps", 10))
	// Output:
	// the quick
	// brown fox
	// jumps
}
