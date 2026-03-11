package maputil_test

import (
	"fmt"
	"slices"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/maputil"
)

func ExampleKeys() {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := maputil.Keys(m)
	slices.Sort(keys)
	fmt.Println(keys)
	// Output:
	// [a b c]
}

func ExampleMerge() {
	base := map[string]int{"a": 1, "b": 2}
	override := map[string]int{"b": 3, "c": 4}
	result := maputil.Merge(base, override)

	keys := maputil.Keys(result)
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Printf("%s:%d ", k, result[k])
	}
	fmt.Println()
	// Output:
	// a:1 b:3 c:4
}

func ExamplePick() {
	m := map[string]int{"name": 1, "age": 2, "email": 3}
	sub := maputil.Pick(m, "name", "email")

	keys := maputil.Keys(sub)
	slices.Sort(keys)
	fmt.Println(keys)
	// Output:
	// [email name]
}

func ExampleMapValues() {
	m := map[string]string{"greeting": "hello", "farewell": "goodbye"}
	upper := maputil.MapValues(m, strings.ToUpper)

	keys := maputil.Keys(upper)
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Printf("%s:%s ", k, upper[k])
	}
	fmt.Println()
	// Output:
	// farewell:GOODBYE greeting:HELLO
}
