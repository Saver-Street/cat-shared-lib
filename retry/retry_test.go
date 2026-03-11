package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

var errTemp = errors.New("temporary error")

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	var calls int
	err := Do(context.Background(), Config{MaxAttempts: 3}, func(ctx context.Context) error {
		calls++
		return nil
	})
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, calls, 1)
}

func TestDo_SuccessOnSecondAttempt(t *testing.T) {
	var calls int
	err := Do(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(ctx context.Context) error {
		calls++
		if calls < 2 {
			return errTemp
		}
		return nil
	})
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, calls, 2)
}

func TestDo_AllAttemptsExhausted(t *testing.T) {
	var calls int
	err := Do(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(ctx context.Context) error {
		calls++
		return errTemp
	})
	testkit.AssertErrorIs(t, err, errTemp)
	testkit.AssertEqual(t, calls, 3)
}

func TestDo_Defaults(t *testing.T) {
	cfg := Config{}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.MaxAttempts, 3)
	testkit.AssertEqual(t, cfg.InitialDelay, 500*time.Millisecond)
	testkit.AssertEqual(t, cfg.MaxDelay, 30*time.Second)
	testkit.AssertEqual(t, cfg.Multiplier, 2.0)
	testkit.AssertEqual(t, cfg.JitterFraction, 0.0)
}

func TestDo_InvalidJitterFraction_DefaultsApplied(t *testing.T) {
	cfg := Config{JitterFraction: -1}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.JitterFraction, 0.25)

	cfg = Config{JitterFraction: 1.5}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.JitterFraction, 0.25)
}

func TestDo_ContextCancelledBeforeFirstAttempt(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Do(ctx, Config{MaxAttempts: 3}, func(ctx context.Context) error {
		t.Fatal("should not be called")
		return nil
	})
	testkit.AssertErrorIs(t, err, context.Canceled)
}

func TestDo_ContextCancelledDuringSleep(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls atomic.Int32

	go func() {
		// Cancel after the first attempt starts.
		for calls.Load() < 1 {
			time.Sleep(time.Millisecond)
		}
		cancel()
	}()

	err := Do(ctx, Config{MaxAttempts: 5, InitialDelay: 5 * time.Second}, func(ctx context.Context) error {
		calls.Add(1)
		return errTemp
	})

	testkit.AssertErrorIs(t, err, context.Canceled)
	testkit.AssertErrorIs(t, err, errTemp)
}

func TestDo_RetryIf_NonRetryableError(t *testing.T) {
	permanent := errors.New("permanent error")
	var calls int

	err := Do(context.Background(), Config{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		RetryIf: func(err error) bool {
			return !errors.Is(err, permanent)
		},
	}, func(ctx context.Context) error {
		calls++
		return permanent
	})

	testkit.AssertEqual(t, calls, 1)
	testkit.AssertErrorIs(t, err, permanent)
}

func TestDo_RetryIf_RetryableError(t *testing.T) {
	retryable := errors.New("retryable")
	var calls int

	err := Do(context.Background(), Config{
		MaxAttempts:  3,
		InitialDelay: time.Millisecond,
		RetryIf: func(err error) bool {
			return errors.Is(err, retryable)
		},
	}, func(ctx context.Context) error {
		calls++
		return retryable
	})

	testkit.AssertEqual(t, calls, 3)
	testkit.AssertErrorIs(t, err, retryable)
}

func TestDo_SingleAttempt(t *testing.T) {
	err := Do(context.Background(), Config{MaxAttempts: 1}, func(ctx context.Context) error {
		return errTemp
	})
	testkit.AssertErrorIs(t, err, errTemp)
}

func TestCalcDelay_ExponentialBackoff(t *testing.T) {
	cfg := Config{
		InitialDelay:   100 * time.Millisecond,
		Multiplier:     2.0,
		MaxDelay:       10 * time.Second,
		JitterFraction: 0,
	}

	expected := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		400 * time.Millisecond,
		800 * time.Millisecond,
	}

	for i, want := range expected {
		got := calcDelay(cfg, i)
		testkit.AssertEqual(t, got, want)
	}
}

func TestCalcDelay_CappedAtMaxDelay(t *testing.T) {
	cfg := Config{
		InitialDelay:   time.Second,
		Multiplier:     10.0,
		MaxDelay:       5 * time.Second,
		JitterFraction: 0,
	}

	got := calcDelay(cfg, 5) // Would be 100000s without cap.
	testkit.AssertEqual(t, got, 5*time.Second)
}

func TestCalcDelay_WithJitter(t *testing.T) {
	cfg := Config{
		InitialDelay:   time.Second,
		Multiplier:     2.0,
		MaxDelay:       time.Minute,
		JitterFraction: 0.5,
	}

	// Run multiple times to check jitter produces some variation.
	seen := make(map[time.Duration]bool)
	for i := 0; i < 100; i++ {
		d := calcDelay(cfg, 0)
		seen[d] = true
		// Should be between 500ms and 1500ms (1s ± 50%).
		if d < 500*time.Millisecond || d > 1500*time.Millisecond {
			t.Errorf("delay out of jitter range: %v", d)
		}
	}
	testkit.AssertTrue(t, len(seen) >= 2)
}

func TestDo_ContextCancelledBeforeAttempt_WithPreviousError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var calls int

	err := Do(ctx, Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(ctx context.Context) error {
		calls++
		if calls == 1 {
			cancel()
			return errTemp
		}
		return nil
	})

	testkit.AssertErrorIs(t, err, context.Canceled)
}

func TestCalcDelay_NegativeInitialDelay(t *testing.T) {
	// Call calcDelay directly (bypassing defaults) with a negative InitialDelay
	// to exercise the defensive delay < 0 guard.
	cfg := Config{
		InitialDelay:   -1,
		Multiplier:     1.0,
		MaxDelay:       time.Second,
		JitterFraction: 0,
	}
	d := calcDelay(cfg, 0)
	testkit.AssertEqual(t, d, time.Duration(0))
}

func TestDo_ContextExpiredBetweenAttempts(t *testing.T) {
	// Exercise the path where ctx.Err() is checked at the top of a loop
	// iteration with a non-nil lastErr from a previous attempt.
	// With nanosecond delays, the timer and ctx.Done() race in the select;
	// repeating gives the timer a chance to win so we reach the ctx.Err()
	// check at the top of the next iteration.
	for i := 0; i < 200; i++ {
		ctx, cancel := context.WithCancel(context.Background())
		calls := 0
		err := Do(ctx, Config{
			MaxAttempts:  3,
			InitialDelay: time.Nanosecond,
			MaxDelay:     time.Nanosecond,
		}, func(ctx context.Context) error {
			calls++
			if calls == 2 {
				cancel()
			}
			return errTemp
		})
		if err == nil {
			t.Fatal("expected error")
		}
		testkit.AssertErrorIs(t, err, errTemp)
	}
}

func BenchmarkDo_NoRetry(b *testing.B) {
	ctx := context.Background()
	cfg := Config{MaxAttempts: 1}
	b.ResetTimer()
	for b.Loop() {
		Do(ctx, cfg, func(ctx context.Context) error { return nil })
	}
}

func TestSimple_Success(t *testing.T) {
	calls := 0
	err := Simple(context.Background(), 3, func(_ context.Context) error {
		calls++
		if calls < 2 {
			return errors.New("transient")
		}
		return nil
	})
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, calls, 2)
}

func TestSimple_Exhausted(t *testing.T) {
	err := Simple(context.Background(), 2, func(_ context.Context) error {
		return errors.New("fail")
	})
	testkit.AssertError(t, err)
	testkit.AssertErrorContains(t, err, "fail")
}
