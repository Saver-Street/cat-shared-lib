package jsonutil_test

import (
"fmt"

"github.com/Saver-Street/cat-shared-lib/jsonutil"
)

func ExamplePretty() {
out, _ := jsonutil.Pretty([]byte(`{"name":"alice","age":30}`))
fmt.Println(string(out))
// Output:
// {
//   "name": "alice",
//   "age": 30
// }
}

func ExampleCompact() {
out, _ := jsonutil.Compact([]byte("{\n  \"a\": 1\n}"))
fmt.Println(string(out))
// Output:
// {"a":1}
}

func ExampleMerge() {
a := []byte(`{"name":"alice"}`)
b := []byte(`{"age":30}`)
out, _ := jsonutil.Merge(a, b)
fmt.Println(string(out))
// Output:
// {"age":30,"name":"alice"}
}

func ExampleGetPath() {
data := []byte(`{"user":{"name":"alice"}}`)
v, _ := jsonutil.GetPath(data, "user.name")
fmt.Println(v)
// Output:
// alice
}
