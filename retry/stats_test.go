package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestDoWithStats_Success(t *testing.T) {
	r := DoWithStats(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(_ context.Context) error {
		return nil
	})
	testkit.AssertTrue(t, r.OK())
	testkit.AssertEqual(t, r.Attempts, 1)
	testkit.AssertNoError(t, r.Err)
	testkit.AssertTrue(t, r.Duration >= 0)
}

func TestDoWithStats_AllFail(t *testing.T) {
	boom := errors.New("boom")
	r := DoWithStats(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, func(_ context.Context) error {
		return boom
	})
	testkit.AssertTrue(t, !r.OK())
	testkit.AssertEqual(t, r.Attempts, 3)
	testkit.AssertError(t, r.Err)
}

func TestDoWithStats_SucceedsOnRetry(t *testing.T) {
	count := 0
	r := DoWithStats(context.Background(), Config{MaxAttempts: 5, InitialDelay: time.Millisecond}, func(_ context.Context) error {
		count++
		if count < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	testkit.AssertTrue(t, r.OK())
	testkit.AssertEqual(t, r.Attempts, 3)
}

func TestDoWithStats_PermanentError(t *testing.T) {
	r := DoWithStats(context.Background(), Config{MaxAttempts: 5, InitialDelay: time.Millisecond}, func(_ context.Context) error {
		return Permanent(errors.New("fatal"))
	})
	testkit.AssertTrue(t, !r.OK())
	testkit.AssertEqual(t, r.Attempts, 1)
	testkit.AssertContains(t, r.Err.Error(), "fatal")
}

func TestDoWithStats_RetryIf(t *testing.T) {
	r := DoWithStats(context.Background(), Config{
		MaxAttempts:  5,
		InitialDelay: time.Millisecond,
		RetryIf:      func(err error) bool { return err.Error() == "retry me" },
	}, func(_ context.Context) error {
		return errors.New("do not retry")
	})
	testkit.AssertEqual(t, r.Attempts, 1)
	testkit.AssertError(t, r.Err)
}

func TestDoWithStats_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r := DoWithStats(ctx, Config{MaxAttempts: 5}, func(_ context.Context) error {
		return errors.New("should not reach")
	})
	testkit.AssertEqual(t, r.Attempts, 0)
	testkit.AssertError(t, r.Err)
}

func TestDoWithStats_ContextCancelledDuringRetry(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	r := DoWithStats(ctx, Config{MaxAttempts: 100, InitialDelay: 100 * time.Millisecond}, func(_ context.Context) error {
		return errors.New("fail")
	})
	testkit.AssertTrue(t, r.Attempts >= 1)
	testkit.AssertError(t, r.Err)
}

func TestDoWithStats_Duration(t *testing.T) {
	r := DoWithStats(context.Background(), Config{MaxAttempts: 2, InitialDelay: 50 * time.Millisecond}, func(_ context.Context) error {
		return errors.New("fail")
	})
	testkit.AssertTrue(t, r.Duration >= 40*time.Millisecond)
}

func TestResult_OK(t *testing.T) {
	testkit.AssertTrue(t, Result{}.OK())
	testkit.AssertTrue(t, !Result{Err: errors.New("x")}.OK())
}

func TestOnRetry(t *testing.T) {
	var retries []int
	count := 0
	fn := OnRetry(func(_ context.Context) error {
		count++
		if count < 3 {
			return errors.New("fail")
		}
		return nil
	}, func(attempt int, _ error) {
		retries = append(retries, attempt)
	})

	err := Do(context.Background(), Config{MaxAttempts: 5, InitialDelay: time.Millisecond}, fn)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, len(retries), 2) // hook called before attempts 2 and 3
}

func TestOnRetry_NilHook(t *testing.T) {
	count := 0
	fn := OnRetry(func(_ context.Context) error {
		count++
		return nil
	}, nil)

	err := Do(context.Background(), Config{MaxAttempts: 3, InitialDelay: time.Millisecond}, fn)
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, count, 1)
}

func BenchmarkDoWithStats(b *testing.B) {
	cfg := Config{MaxAttempts: 1, InitialDelay: time.Millisecond}
	fn := func(_ context.Context) error { return nil }
	for b.Loop() {
		DoWithStats(context.Background(), cfg, fn)
	}
}
