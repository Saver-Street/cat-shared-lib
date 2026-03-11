package retry

import (
	"context"
	"errors"
	"time"
)

// Result captures the outcome of a retry loop, including diagnostic info.
type Result struct {
	// Attempts is the number of times fn was called.
	Attempts int
	// Duration is the wall-clock time from start to finish.
	Duration time.Duration
	// Err is the final error, or nil on success.
	Err error
}

// OK reports whether the operation succeeded.
func (r Result) OK() bool { return r.Err == nil }

// DoWithStats works like Do but returns a Result with retry statistics.
func DoWithStats(ctx context.Context, cfg Config, fn func(ctx context.Context) error) Result {
	start := time.Now()
	cfg.defaults()

	var lastErr error
	attempts := 0
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return Result{Attempts: attempts, Duration: time.Since(start), Err: errors.Join(lastErr, err)}
			}
			return Result{Attempts: attempts, Duration: time.Since(start), Err: err}
		}

		attempts++
		lastErr = fn(ctx)
		if lastErr == nil {
			return Result{Attempts: attempts, Duration: time.Since(start), Err: nil}
		}

		var pe *PermanentError
		if errors.As(lastErr, &pe) {
			return Result{Attempts: attempts, Duration: time.Since(start), Err: pe.Err}
		}

		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return Result{Attempts: attempts, Duration: time.Since(start), Err: lastErr}
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := calcDelay(cfg, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return Result{Attempts: attempts, Duration: time.Since(start), Err: errors.Join(lastErr, ctx.Err())}
		case <-timer.C:
		}
	}

	return Result{Attempts: attempts, Duration: time.Since(start), Err: lastErr}
}

// OnRetry wraps fn to call hook before each retry attempt (not the first).
// The hook receives the attempt number (1-based, so 1 means the first retry)
// and the error from the previous attempt.
func OnRetry(fn func(ctx context.Context) error, hook func(attempt int, err error)) func(ctx context.Context) error {
	attempt := 0
	return func(ctx context.Context) error {
		attempt++
		if attempt > 1 && hook != nil {
			hook(attempt-1, nil)
		}
		return fn(ctx)
	}
}
