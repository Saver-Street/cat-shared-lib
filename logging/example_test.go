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

func ExampleFromContext() {
	ctx := context.Background()
	logger := logging.FromContext(ctx)
	fmt.Println(logger != nil)
	// Output:
	// true
}

func ExampleWith() {
	logger := logging.New()
	ctx := logging.WithContext(context.Background(), logger)
	child := logging.With(ctx, "component", "auth")
	fmt.Println(child != nil)
	// Output:
	// true
}

func ExampleRequestID() {
	attr := logging.RequestID("req-123")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// request_id req-123
}

func ExampleUserID() {
	attr := logging.UserID("user-456")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// user_id user-456
}

func ExampleVersion() {
	attr := logging.Version("1.2.3")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// version 1.2.3
}

func ExampleComponent() {
	attr := logging.Component("middleware")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// component middleware
}

func ExampleTraceID() {
	attr := logging.TraceID("abc-trace-id")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// trace_id abc-trace-id
}

func ExampleWithWriter() {
	var buf bytes.Buffer
	logger := logging.New(logging.WithWriter(&buf))
	logger.Info("buffered")
	fmt.Println(strings.Contains(buf.String(), "buffered"))
	// Output:
	// true
}

func ExampleWithLevel() {
	var buf bytes.Buffer
	logger := logging.New(logging.WithWriter(&buf), logging.WithLevel(slog.LevelError))
	logger.Info("should be filtered")
	fmt.Println(buf.Len() == 0)
	// Output:
	// true
}

func ExampleWithFormat() {
	var buf bytes.Buffer
	logger := logging.New(logging.WithFormat("json"), logging.WithWriter(&buf))
	logger.Info("test")
	fmt.Println(strings.Contains(buf.String(), "{"))
	// Output:
	// true
}

func ExampleWithSource() {
	var buf bytes.Buffer
	logger := logging.New(logging.WithSource(), logging.WithFormat("json"), logging.WithWriter(&buf))
	logger.Info("test")
	fmt.Println(strings.Contains(buf.String(), "source"))
	// Output:
	// true
}

func ExampleWithAttrs() {
	var buf bytes.Buffer
	logger := logging.New(
		logging.WithAttrs(logging.Service("api")),
		logging.WithFormat("json"),
		logging.WithWriter(&buf),
	)
	logger.Info("boot")
	fmt.Println(strings.Contains(buf.String(), "api"))
	// Output:
	// true
}
