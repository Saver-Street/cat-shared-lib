package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/retry"
)

func ExampleDo() {
	attempts := 0
	err := retry.Do(context.Background(), retry.Config{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
	}, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure")
		}
		return nil
	})
	fmt.Println(err)
	fmt.Println(attempts)
	// Output:
	// <nil>
	// 3
}

func ExampleDo_permanent() {
	err := retry.Do(context.Background(), retry.Config{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
	}, func(ctx context.Context) error {
		return errors.New("always fails")
	})
	fmt.Println(err)
	// Output:
	// always fails
}

func ExampleDoWithStats() {
	result := retry.DoWithStats(context.Background(), retry.Config{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
	}, func(_ context.Context) error {
		return nil
	})
	fmt.Printf("ok=%t attempts=%d\n", result.OK(), result.Attempts)
	// Output:
	// ok=true attempts=1
}

func ExampleResult_OK() {
	r := retry.Result{Attempts: 1}
	fmt.Println(r.OK())
	// Output:
	// true
}
