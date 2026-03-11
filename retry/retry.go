// Package retry provides exponential backoff with jitter for retrying
// fallible operations. It supports context cancellation and configurable
// retry conditions.
package retry

import (
	"context"
	"errors"
	"math"
	"math/rand/v2"
	"time"
)

// Config configures retry behaviour.
type Config struct {
	// MaxAttempts is the maximum number of attempts (including the first).
	// Default: 3. Set to 1 for no retries.
	MaxAttempts int
	// InitialDelay is the base delay before the first retry. Default: 500ms.
	InitialDelay time.Duration
	// MaxDelay caps the delay between retries. Default: 30s.
	MaxDelay time.Duration
	// Multiplier is the backoff factor applied after each attempt. Default: 2.0.
	Multiplier float64
	// JitterFraction is the fraction of the delay to randomise (0.0–1.0).
	// Default: 0.25 (±25%).
	JitterFraction float64
	// RetryIf is an optional predicate that decides whether to retry for a
	// given error. If nil, all non-nil errors are retried.
	RetryIf func(error) bool
}

func (c *Config) defaults() {
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 3
	}
	if c.InitialDelay <= 0 {
		c.InitialDelay = 500 * time.Millisecond
	}
	if c.MaxDelay <= 0 {
		c.MaxDelay = 30 * time.Second
	}
	if c.Multiplier <= 0 {
		c.Multiplier = 2.0
	}
	if c.JitterFraction < 0 || c.JitterFraction > 1 {
		c.JitterFraction = 0.25
	}
}

// Do executes fn up to MaxAttempts times, applying exponential backoff with
// jitter between retries. It returns the first nil error or the last error
// after all attempts are exhausted. If ctx is cancelled, it returns ctx.Err()
// wrapped with the last operation error.
func Do(ctx context.Context, cfg Config, fn func(ctx context.Context) error) error {
	cfg.defaults()

	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return errors.Join(lastErr, err)
			}
			return err
		}

		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}

		// Stop immediately on permanent errors.
		var pe *PermanentError
		if errors.As(lastErr, &pe) {
			return pe.Err
		}

		// Check if this error is retryable.
		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return lastErr
		}

		// Don't sleep after the last attempt.
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := calcDelay(cfg, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return errors.Join(lastErr, ctx.Err())
		case <-timer.C:
		}
	}

	return lastErr
}

func calcDelay(cfg Config, attempt int) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}

	if cfg.JitterFraction > 0 {
		jitter := delay * cfg.JitterFraction
		delay = delay - jitter + rand.Float64()*2*jitter //nolint:gosec
	}

	if delay < 0 {
		delay = 0
	}
	return time.Duration(delay)
}

// Simple retries fn up to maxAttempts times with default backoff settings.
// It is a shorthand for Do with only MaxAttempts configured.
func Simple(ctx context.Context, maxAttempts int, fn func(ctx context.Context) error) error {
	return Do(ctx, Config{MaxAttempts: maxAttempts}, fn)
}

// WithTimeout retries fn with the given config but limits the total elapsed
// time to timeout. It derives a child context with the deadline and delegates
// to Do.
func WithTimeout(ctx context.Context, timeout time.Duration, cfg Config, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return Do(ctx, cfg, fn)
}

// Delay calculates the backoff delay for the given attempt number (0-based)
// using the provided Config. This is useful for logging or implementing
// custom retry loops without using Do directly.
func Delay(cfg Config, attempt int) time.Duration {
	cfg.defaults()
	return calcDelay(cfg, attempt)
}

// PermanentError wraps an error to signal that it should not be retried.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

// Permanent wraps err so that Do and DoWithResult stop retrying immediately.
func Permanent(err error) error {
	if err == nil {
		return nil
	}
	return &PermanentError{Err: err}
}

// IsPermanent reports whether err (or any error in its chain) is a PermanentError.
func IsPermanent(err error) bool {
	var pe *PermanentError
	return errors.As(err, &pe)
}

// DoWithResult executes fn up to MaxAttempts times, returning the result
// value on success. It follows the same backoff and cancellation semantics
// as Do. If fn returns a [PermanentError], retries stop immediately.
func DoWithResult[T any](ctx context.Context, cfg Config, fn func(ctx context.Context) (T, error)) (T, error) {
	cfg.defaults()

	var (
		lastErr error
		zero    T
	)
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			if lastErr != nil {
				return zero, errors.Join(lastErr, err)
			}
			return zero, err
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Stop immediately on permanent errors.
		var pe *PermanentError
		if errors.As(err, &pe) {
			return zero, pe.Err
		}

		if cfg.RetryIf != nil && !cfg.RetryIf(lastErr) {
			return zero, lastErr
		}

		if attempt == cfg.MaxAttempts-1 {
			break
		}

		delay := calcDelay(cfg, attempt)
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, errors.Join(lastErr, ctx.Err())
		case <-timer.C:
		}
	}

	return zero, lastErr
}
