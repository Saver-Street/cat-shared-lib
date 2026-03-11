package timeout_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/timeout"
)

func ExampleDo() {
	result, err := timeout.Do(context.Background(), time.Second, func(_ context.Context) (string, error) {
		return "done", nil
	})
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output: done
}

func ExampleAfter() {
	ch := timeout.After(context.Background(), func(_ context.Context) int {
		return 42
	})
	fmt.Println(<-ch)
	// Output: 42
}

func ExampleRace() {
	fast := func(_ context.Context) (string, error) {
		return "winner", nil
	}
	slow := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(time.Hour):
			return "slow", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	result, err := timeout.Race(context.Background(), fast, slow)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println(result)
	// Output: winner
}
