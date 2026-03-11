package retry

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"
)

func FuzzCalcDelay(f *testing.F) {
	f.Add(int64(500*time.Millisecond), int64(30*time.Second), 2.0, 0.25, 0)
	f.Add(int64(0), int64(0), 0.0, 0.0, 0)
	f.Add(int64(time.Nanosecond), int64(time.Hour), 10.0, 1.0, 5)
	f.Add(int64(-1), int64(-1), -1.0, -0.5, 100)
	f.Add(int64(time.Second), int64(time.Minute), math.MaxFloat64, 0.5, 50)

	f.Fuzz(func(t *testing.T, initial, maxD int64, mult, jitter float64, attempt int) {
		if math.IsNaN(mult) || math.IsInf(mult, 0) {
			t.Skip()
		}
		if math.IsNaN(jitter) || math.IsInf(jitter, 0) {
			t.Skip()
		}

		cfg := Config{
			InitialDelay:   time.Duration(initial),
			MaxDelay:       time.Duration(maxD),
			Multiplier:     mult,
			JitterFraction: jitter,
		}
		cfg.defaults()

		d := calcDelay(cfg, attempt)
		if d < 0 {
			t.Errorf("calcDelay returned negative: %v", d)
		}
	})
}

func FuzzDo(f *testing.F) {
	f.Add(1, int64(time.Millisecond), 2.0, true)
	f.Add(3, int64(time.Nanosecond), 1.5, false)
	f.Add(0, int64(0), 0.0, true)
	f.Add(10, int64(time.Microsecond), 0.5, false)

	f.Fuzz(func(t *testing.T, maxAttempts int, initialDelay int64, mult float64, succeed bool) {
		if math.IsNaN(mult) || math.IsInf(mult, 0) {
			t.Skip()
		}
		// Clamp to avoid slow tests.
		if maxAttempts > 5 {
			maxAttempts = 5
		}
		if initialDelay > int64(time.Millisecond) {
			initialDelay = int64(time.Millisecond)
		}

		cfg := Config{
			MaxAttempts:  maxAttempts,
			InitialDelay: time.Duration(initialDelay),
			MaxDelay:     time.Millisecond,
			Multiplier:   mult,
		}

		calls := 0
		errBoom := errors.New("boom")
		_ = Do(context.Background(), cfg, func(_ context.Context) error {
			calls++
			if succeed {
				return nil
			}
			return errBoom
		})
		_ = calls
	})
}
