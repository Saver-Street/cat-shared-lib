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
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
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
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if calls != 2 {
		t.Errorf("expected 2 calls, got %d", calls)
	}
}

func TestDo_AllAttemptsExhausted(t *testing.T) {
	var calls int
	err := Do(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(ctx context.Context) error {
		calls++
		return errTemp
	})
	testkit.AssertErrorIs(t, err, errTemp)
	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
}

func TestDo_Defaults(t *testing.T) {
	cfg := Config{}
	cfg.defaults()
	if cfg.MaxAttempts != 3 {
		t.Errorf("expected 3 max attempts, got %d", cfg.MaxAttempts)
	}
	if cfg.InitialDelay != 500*time.Millisecond {
		t.Errorf("expected 500ms initial delay, got %v", cfg.InitialDelay)
	}
	if cfg.MaxDelay != 30*time.Second {
		t.Errorf("expected 30s max delay, got %v", cfg.MaxDelay)
	}
	if cfg.Multiplier != 2.0 {
		t.Errorf("expected 2.0 multiplier, got %f", cfg.Multiplier)
	}
	if cfg.JitterFraction != 0 {
		t.Errorf("expected JitterFraction = 0 after defaults(), got %f", cfg.JitterFraction)
	}
}

func TestDo_InvalidJitterFraction_DefaultsApplied(t *testing.T) {
	cfg := Config{JitterFraction: -1}
	cfg.defaults()
	if cfg.JitterFraction != 0.25 {
		t.Errorf("expected 0.25 jitter for negative input, got %f", cfg.JitterFraction)
	}

	cfg = Config{JitterFraction: 1.5}
	cfg.defaults()
	if cfg.JitterFraction != 0.25 {
		t.Errorf("expected 0.25 jitter for >1 input, got %f", cfg.JitterFraction)
	}
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

	if calls != 1 {
		t.Errorf("expected 1 call for non-retryable error, got %d", calls)
	}
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

	if calls != 3 {
		t.Errorf("expected 3 calls, got %d", calls)
	}
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
		if got != want {
			t.Errorf("attempt %d: expected %v, got %v", i, want, got)
		}
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
	if got != 5*time.Second {
		t.Errorf("expected capped at 5s, got %v", got)
	}
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
	if len(seen) < 2 {
		t.Error("jitter should produce variable delays")
	}
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
	if d != 0 {
		t.Errorf("expected 0 for negative delay, got %v", d)
	}
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
	for i := 0; i < b.N; i++ {
		Do(ctx, cfg, func(ctx context.Context) error { return nil })
	}
}
