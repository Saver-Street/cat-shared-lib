package convert_test

import (
"fmt"

"github.com/Saver-Street/cat-shared-lib/convert"
)

func ExampleToInt() {
port := convert.ToInt("8080", 3000)
fallback := convert.ToInt("not-a-number", 3000)
fmt.Println(port)
fmt.Println(fallback)
// Output:
// 8080
// 3000
}

func ExampleToBool() {
debug := convert.ToBool("yes", false)
unknown := convert.ToBool("maybe", false)
fmt.Println(debug)
fmt.Println(unknown)
// Output:
// true
// false
}

func ExamplePtr() {
p := convert.Ptr("hello")
fmt.Println(*p)
// Output:
// hello
}

func ExampleDeref() {
var p *int
fmt.Println(convert.Deref(p))
v := 42
fmt.Println(convert.Deref(&v))
// Output:
// 0
// 42
}
