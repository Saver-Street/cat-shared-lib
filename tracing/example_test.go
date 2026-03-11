package tracing_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/tracing"
)

func ExampleNewProvider() {
	ctx := context.Background()
	tp, err := tracing.NewProvider(ctx, tracing.Config{
		ServiceName:    "example-service",
		ServiceVersion: "v1.0.0",
		Exporter:       tracing.ExporterNoop,
	})
	if err != nil {
		panic(err)
	}
	defer tp.Shutdown(ctx)

	fmt.Printf("%T\n", tp)
	// Output:
	// *tracing.Provider
}

func ExampleRecordError() {
	ctx := context.Background()
	tp, _ := tracing.NewProvider(ctx, tracing.Config{
		ServiceName: "demo",
		Exporter:    tracing.ExporterNoop,
	})
	defer tp.Shutdown(ctx)

	tracer := tp.Tracer("example")
	_, span := tracer.Start(ctx, "operation")
	defer span.End()

	// RecordError is a no-op when err is nil.
	tracing.RecordError(span, nil)

	// Record a real error.
	tracing.RecordError(span, errors.New("something failed"))
	fmt.Println("error recorded")
	// Output:
	// error recorded
}

func ExampleTraceID() {
	// Without an active span, TraceID returns "".
	id := tracing.TraceID(context.Background())
	fmt.Printf("empty: %q\n", id)
	// Output:
	// empty: ""
}

func ExampleIsRecording() {
	// Without an active span, IsRecording returns false.
	fmt.Println(tracing.IsRecording(context.Background()))
	// Output:
	// false
}
