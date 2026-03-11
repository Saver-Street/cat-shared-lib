package logging_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Saver-Street/cat-shared-lib/logging"
)

func ExampleNew() {
	logger := logging.New(logging.WithFormat("text"), logging.WithLevel(slog.LevelInfo))
	logger.Info("server started")
	// Outputs a structured log line (format depends on options).
}

func ExampleNew_json() {
	var buf bytes.Buffer
	logger := logging.New(logging.WithFormat("json"), logging.WithWriter(&buf))
	logger.Info("hello")
	fmt.Println(strings.Contains(buf.String(), `"msg":"hello"`))
	// Output:
	// true
}

func ExampleParseLevel() {
	fmt.Println(logging.ParseLevel("debug"))
	fmt.Println(logging.ParseLevel("warn"))
	fmt.Println(logging.ParseLevel("INVALID"))
	// Output:
	// DEBUG
	// WARN
	// INFO
}

func ExampleWithContext() {
	logger := logging.New()
	ctx := logging.WithContext(context.Background(), logger)
	got := logging.FromContext(ctx)
	fmt.Println(got != nil)
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

func ExampleService() {
	attr := logging.Service("catalog")
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// service catalog
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

func ExampleErr() {
	attr := logging.Err(errors.New("disk full"))
	fmt.Println(attr.Key, attr.Value.String())
	// Output:
	// error disk full
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
