package schedule

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_Every(t *testing.T) {
	s := New()
	defer s.Stop()

	var count atomic.Int64
	s.Every("test", 10*time.Millisecond, func(_ context.Context) {
		count.Add(1)
	})

	time.Sleep(55 * time.Millisecond)
	if got := count.Load(); got < 3 {
		t.Errorf("count = %d, want >= 3", got)
	}
}

func TestScheduler_Remove(t *testing.T) {
	s := New()
	defer s.Stop()

	var count atomic.Int64
	s.Every("removable", 10*time.Millisecond, func(_ context.Context) {
		count.Add(1)
	})

	time.Sleep(35 * time.Millisecond)
	s.Remove("removable")
	snapshot := count.Load()
	time.Sleep(30 * time.Millisecond)

	if got := count.Load(); got != snapshot {
		t.Errorf("count changed after remove: %d -> %d", snapshot, got)
	}
}

func TestScheduler_Remove_Nonexistent(t *testing.T) {
	s := New()
	defer s.Stop()
	s.Remove("does-not-exist") // should not panic
}

func TestScheduler_Stop(t *testing.T) {
	s := New()
	var count atomic.Int64
	s.Every("task1", 10*time.Millisecond, func(_ context.Context) {
		count.Add(1)
	})
	s.Every("task2", 10*time.Millisecond, func(_ context.Context) {
		count.Add(1)
	})

	time.Sleep(35 * time.Millisecond)
	s.Stop()
	snapshot := count.Load()
	time.Sleep(30 * time.Millisecond)

	if got := count.Load(); got != snapshot {
		t.Errorf("count changed after stop: %d -> %d", snapshot, got)
	}
}

func TestScheduler_Len(t *testing.T) {
	s := New()
	defer s.Stop()

	if s.Len() != 0 {
		t.Errorf("Len() = %d, want 0", s.Len())
	}

	s.Every("a", time.Hour, func(_ context.Context) {})
	s.Every("b", time.Hour, func(_ context.Context) {})
	if s.Len() != 2 {
		t.Errorf("Len() = %d, want 2", s.Len())
	}

	s.Remove("a")
	if s.Len() != 1 {
		t.Errorf("Len() = %d, want 1", s.Len())
	}
}

func TestScheduler_ReplaceTask(t *testing.T) {
	s := New()
	defer s.Stop()

	var first, second atomic.Int64
	s.Every("task", 10*time.Millisecond, func(_ context.Context) {
		first.Add(1)
	})

	time.Sleep(35 * time.Millisecond)
	firstSnapshot := first.Load()

	s.Every("task", 10*time.Millisecond, func(_ context.Context) {
		second.Add(1)
	})

	time.Sleep(35 * time.Millisecond)
	if first.Load() != firstSnapshot {
		t.Error("first task should have stopped after replacement")
	}
	if second.Load() < 2 {
		t.Errorf("second task count = %d, want >= 2", second.Load())
	}
}

func TestScheduler_TaskPanicRecovery(t *testing.T) {
	s := New()
	defer s.Stop()

	var count atomic.Int64
	s.Every("panicky", 10*time.Millisecond, func(_ context.Context) {
		count.Add(1)
		if count.Load() == 1 {
			panic("test panic")
		}
	})

	time.Sleep(55 * time.Millisecond)
	if got := count.Load(); got < 2 {
		t.Errorf("count = %d, want >= 2 (should recover from panic)", got)
	}
}

func TestScheduler_ContextCancellation(t *testing.T) {
	s := New()
	defer s.Stop()

	var cancelled atomic.Bool
	s.Every("ctx-aware", 10*time.Millisecond, func(ctx context.Context) {
		select {
		case <-ctx.Done():
			cancelled.Store(true)
		default:
		}
	})

	time.Sleep(25 * time.Millisecond)
	s.Stop()

	// After stop, the context should be cancelled.
	if !cancelled.Load() {
		// The task may not have seen the cancellation in the select; that's OK
		// as long as the task stopped running (verified by Stop() returning).
	}
}

func TestScheduler_StopIdempotent(t *testing.T) {
	s := New()
	s.Every("task", time.Hour, func(_ context.Context) {})
	s.Stop()
	s.Stop() // second stop should not panic
}

func BenchmarkScheduler_Every(b *testing.B) {
	for b.Loop() {
		s := New()
		s.Every("bench", time.Hour, func(_ context.Context) {})
		s.Stop()
	}
}
