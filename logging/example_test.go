package logging_test

import (
"bytes"
"context"
"fmt"
"log/slog"
"strings"

"github.com/Saver-Street/cat-shared-lib/logging"
)

func ExampleNew() {
var buf bytes.Buffer
l := logging.New(logging.WithWriter(&buf), logging.WithFormat("text"), logging.WithLevel(slog.LevelInfo))
l.Info("hello", "key", "value")
fmt.Println(strings.Contains(buf.String(), "hello"))
// Output:
// true
}

func ExampleNew_json() {
var buf bytes.Buffer
l := logging.New(logging.WithWriter(&buf), logging.WithFormat("json"))
l.Info("test")
fmt.Println(strings.Contains(buf.String(), `"msg":"test"`))
// Output:
// true
}

func ExampleParseLevel() {
fmt.Println(logging.ParseLevel("debug"))
fmt.Println(logging.ParseLevel("warn"))
fmt.Println(logging.ParseLevel("ERROR"))
// Output:
// DEBUG
// WARN
// ERROR
}

func ExampleWithContext() {
var buf bytes.Buffer
l := logging.New(logging.WithWriter(&buf), logging.WithFormat("text"))
ctx := logging.WithContext(context.Background(), l)
logging.FromContext(ctx).Info("from context")
fmt.Println(strings.Contains(buf.String(), "from context"))
// Output:
// true
}

func ExampleService() {
attr := logging.Service("auth")
fmt.Println(attr.Key)
fmt.Println(attr.Value.String())
// Output:
// service
// auth
}

func ExampleErr() {
attr := logging.Err(fmt.Errorf("oops"))
fmt.Println(attr.Key)
fmt.Println(attr.Value.String())
// Output:
// error
// oops
}

func ExampleErr_nil() {
attr := logging.Err(nil)
fmt.Println(attr.Equal(slog.Attr{}))
// Output:
// true
}
