package worker

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestPool_BasicExecution(t *testing.T) {
	pool := New(2)
	var count atomic.Int64

	for range 10 {
		pool.Submit(func(_ context.Context) error {
			count.Add(1)
			return nil
		})
	}

	errs := pool.Shutdown(context.Background())
	if len(errs) != 0 {
		t.Errorf("got %d errors, want 0", len(errs))
	}
	if got := count.Load(); got != 10 {
		t.Errorf("count = %d, want 10", got)
	}
}

func TestPool_CollectsErrors(t *testing.T) {
	pool := New(2)
	errBad := errors.New("bad")

	pool.Submit(func(_ context.Context) error { return nil })
	pool.Submit(func(_ context.Context) error { return errBad })
	pool.Submit(func(_ context.Context) error { return nil })
	pool.Submit(func(_ context.Context) error { return errBad })

	errs := pool.Shutdown(context.Background())
	if len(errs) != 2 {
		t.Errorf("got %d errors, want 2", len(errs))
	}
}

func TestPool_PanicRecovery(t *testing.T) {
	pool := New(1)
	var count atomic.Int64

	pool.Submit(func(_ context.Context) error {
		panic("test panic")
	})
	pool.Submit(func(_ context.Context) error {
		count.Add(1)
		return nil
	})

	errs := pool.Shutdown(context.Background())
	if len(errs) != 0 {
		t.Errorf("got %d errors, want 0 (panic is logged, not returned)", len(errs))
	}
	if got := count.Load(); got != 1 {
		t.Errorf("count = %d, want 1 (second job should still run)", got)
	}
}

func TestPool_ContextCancellation(t *testing.T) {
	pool := New(1)

	pool.Submit(func(_ context.Context) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	errs := pool.Shutdown(ctx)
	_ = errs // just ensure it completes without deadlock
}

func TestPool_MinWorkers(t *testing.T) {
	pool := New(0) // should default to 1
	var count atomic.Int64

	pool.Submit(func(_ context.Context) error {
		count.Add(1)
		return nil
	})

	pool.Shutdown(context.Background())
	if got := count.Load(); got != 1 {
		t.Errorf("count = %d, want 1", got)
	}
}

func TestPool_ConcurrentWorkers(t *testing.T) {
	pool := New(4)
	var peak atomic.Int64
	var current atomic.Int64

	for range 20 {
		pool.Submit(func(_ context.Context) error {
			n := current.Add(1)
			for {
				old := peak.Load()
				if n <= old || peak.CompareAndSwap(old, n) {
					break
				}
			}
			time.Sleep(10 * time.Millisecond)
			current.Add(-1)
			return nil
		})
	}

	pool.Shutdown(context.Background())
	if got := peak.Load(); got < 2 {
		t.Errorf("peak concurrency = %d, want >= 2", got)
	}
}

func TestPool_JobReceivesContext(t *testing.T) {
	pool := New(1)
	var cancelled atomic.Bool

	pool.Submit(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			cancelled.Store(true)
		default:
		}
		return nil
	})

	pool.Shutdown(context.Background())
	// Context should not be cancelled during normal operation
	if cancelled.Load() {
		t.Error("context was cancelled during normal operation")
	}
}

func TestPool_ShutdownEmpty(t *testing.T) {
	pool := New(2)
	errs := pool.Shutdown(context.Background())
	if len(errs) != 0 {
		t.Errorf("got %d errors, want 0", len(errs))
	}
}

func BenchmarkPool_Submit(b *testing.B) {
	pool := New(4)
	noop := func(_ context.Context) error { return nil }

	for b.Loop() {
		pool.Submit(noop)
	}
	pool.Shutdown(context.Background())
}

func BenchmarkPool_SubmitShutdown(b *testing.B) {
	for b.Loop() {
		pool := New(4)
		for range 100 {
			pool.Submit(func(_ context.Context) error { return nil })
		}
		pool.Shutdown(context.Background())
	}
}
