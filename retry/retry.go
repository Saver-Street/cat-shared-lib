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
		delay = delay - jitter + rand.Float64()*2*jitter
	}

	if delay < 0 {
		delay = 0
	}
	return time.Duration(delay)
}
