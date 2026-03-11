package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func BenchmarkDo_SuccessFirstAttempt(b *testing.B) {
	ctx := context.Background()
	cfg := Config{MaxAttempts: 3}
	for b.Loop() {
		Do(ctx, cfg, func(ctx context.Context) error { return nil })
	}
}

func BenchmarkDo_SuccessAfterRetry(b *testing.B) {
	ctx := context.Background()
	cfg := Config{MaxAttempts: 3, InitialDelay: time.Nanosecond, MaxDelay: time.Nanosecond}
	errTemp := errors.New("temp")
	for b.Loop() {
		attempt := 0
		Do(ctx, cfg, func(ctx context.Context) error {
			attempt++
			if attempt < 2 {
				return errTemp
			}
			return nil
		})
	}
}

func BenchmarkCalcDelay(b *testing.B) {
	cfg := Config{
		InitialDelay:   500,
		MaxDelay:       30000,
		Multiplier:     2.0,
		JitterFraction: 0.25,
	}
	for b.Loop() {
		calcDelay(cfg, 5)
	}
}

func BenchmarkCalcDelay_NoJitter(b *testing.B) {
	cfg := Config{
		InitialDelay:   500,
		MaxDelay:       30000,
		Multiplier:     2.0,
		JitterFraction: 0,
	}
	for b.Loop() {
		calcDelay(cfg, 5)
	}
}
