package timeout

import (
	"context"
	"fmt"
	"time"
)

// Do runs fn with a timeout. If fn completes before the timeout, its result
// is returned. If the timeout is reached, a timeout error is returned.
// The context passed to fn is cancelled when the timeout expires.
func Do[T any](ctx context.Context, d time.Duration, fn func(context.Context) (T, error)) (T, error) {
	tctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()

	type result struct {
		val T
		err error
	}

	ch := make(chan result, 1)
	go func() {
		v, err := fn(tctx)
		ch <- result{val: v, err: err}
	}()

	select {
	case r := <-ch:
		return r.val, r.err
	case <-tctx.Done():
		var zero T
		if ctx.Err() != nil {
			return zero, ctx.Err()
		}
		return zero, fmt.Errorf("timeout: operation exceeded %s deadline", d)
	}
}

// DoSimple runs fn with a timeout. If fn doesn't complete in time, an error
// is returned. For functions that don't return a value.
func DoSimple(ctx context.Context, d time.Duration, fn func(context.Context) error) error {
	_, err := Do(ctx, d, func(c context.Context) (struct{}, error) {
		return struct{}{}, fn(c)
	})
	return err
}

// After returns a channel that receives a value after the function completes
// or the context is cancelled.
func After[T any](ctx context.Context, fn func(context.Context) T) <-chan T {
	ch := make(chan T, 1)
	go func() {
		ch <- fn(ctx)
	}()
	return ch
}

// Race runs multiple functions concurrently and returns the result of the
// first one to complete. If the context is cancelled before any completes,
// a context error is returned. All other functions' contexts are cancelled
// once one completes.
func Race[T any](ctx context.Context, fns ...func(context.Context) (T, error)) (T, error) {
	if len(fns) == 0 {
		var zero T
		return zero, fmt.Errorf("timeout: Race called with no functions")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		val T
		err error
	}

	ch := make(chan result, len(fns))
	for _, fn := range fns {
		go func() {
			v, err := fn(ctx)
			ch <- result{val: v, err: err}
		}()
	}

	select {
	case r := <-ch:
		return r.val, r.err
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	}
}
