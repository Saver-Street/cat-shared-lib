package timeout

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// ---------- Do ----------

func TestDo_FastFunction(t *testing.T) {
	result, err := Do(context.Background(), time.Second, func(_ context.Context) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 42 {
		t.Fatalf("got %d, want 42", result)
	}
}

func TestDo_SlowFunction_Timeout(t *testing.T) {
	_, err := Do(context.Background(), 50*time.Millisecond, func(ctx context.Context) (int, error) {
		select {
		case <-time.After(5 * time.Second):
			return 1, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !strings.Contains(err.Error(), "timeout") && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDo_AlreadyCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Do(ctx, time.Second, func(ctx context.Context) (string, error) {
		return "hello", nil
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestDo_FunctionReturnsError(t *testing.T) {
	wantErr := errors.New("fn failed")
	_, err := Do(context.Background(), time.Second, func(_ context.Context) (int, error) {
		return 0, wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

func TestDo_ZeroDurationTimesOut(t *testing.T) {
	_, err := Do(context.Background(), 0, func(ctx context.Context) (int, error) {
		select {
		case <-time.After(time.Second):
			return 1, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	if err == nil {
		t.Fatal("expected timeout error with zero duration")
	}
}

func TestDo_StringResult(t *testing.T) {
	result, err := Do(context.Background(), time.Second, func(_ context.Context) (string, error) {
		return "hello", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Fatalf("got %q, want %q", result, "hello")
	}
}

// ---------- DoSimple ----------

func TestDoSimple_Success(t *testing.T) {
	err := DoSimple(context.Background(), time.Second, func(_ context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoSimple_Timeout(t *testing.T) {
	err := DoSimple(context.Background(), 50*time.Millisecond, func(ctx context.Context) error {
		select {
		case <-time.After(5 * time.Second):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestDoSimple_FunctionReturnsError(t *testing.T) {
	wantErr := errors.New("simple fail")
	err := DoSimple(context.Background(), time.Second, func(_ context.Context) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("got %v, want %v", err, wantErr)
	}
}

func TestDoSimple_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := DoSimple(ctx, time.Second, func(ctx context.Context) error {
		return nil
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// ---------- After ----------

func TestAfter_CompletesSuccessfully(t *testing.T) {
	ch := After(context.Background(), func(_ context.Context) int {
		return 99
	})
	select {
	case v := <-ch:
		if v != 99 {
			t.Fatalf("got %d, want 99", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for After")
	}
}

func TestAfter_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch := After(ctx, func(ctx context.Context) string {
		<-ctx.Done()
		return "cancelled"
	})
	select {
	case v := <-ch:
		if v != "cancelled" {
			t.Fatalf("got %q, want %q", v, "cancelled")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for After with cancelled context")
	}
}

func TestAfter_SlowFunction(t *testing.T) {
	ch := After(context.Background(), func(_ context.Context) int {
		time.Sleep(50 * time.Millisecond)
		return 7
	})
	select {
	case v := <-ch:
		if v != 7 {
			t.Fatalf("got %d, want 7", v)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

// ---------- Race ----------

func TestRace_FirstOneWins(t *testing.T) {
	fast := func(_ context.Context) (string, error) {
		return "fast", nil
	}
	slow := func(ctx context.Context) (string, error) {
		select {
		case <-time.After(5 * time.Second):
			return "slow", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
	result, err := Race(context.Background(), fast, slow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "fast" {
		t.Fatalf("got %q, want %q", result, "fast")
	}
}

func TestRace_AllFailing(t *testing.T) {
	fail := func(_ context.Context) (int, error) {
		return 0, errors.New("fail")
	}
	_, err := Race(context.Background(), fail, fail, fail)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRace_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Race(ctx, func(ctx context.Context) (int, error) {
		select {
		case <-time.After(5 * time.Second):
			return 1, nil
		case <-ctx.Done():
			return 0, ctx.Err()
		}
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRace_NoFunctions(t *testing.T) {
	_, err := Race[int](context.Background())
	if err == nil {
		t.Fatal("expected error for empty Race")
	}
}

func TestRace_SingleFunction(t *testing.T) {
	result, err := Race(context.Background(), func(_ context.Context) (int, error) {
		return 100, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != 100 {
		t.Fatalf("got %d, want 100", result)
	}
}

// ---------- Benchmarks ----------

func BenchmarkDo_FastPath(b *testing.B) {
	ctx := context.Background()
	for b.Loop() {
		_, _ = Do(ctx, time.Second, func(_ context.Context) (int, error) {
			return 42, nil
		})
	}
}

func BenchmarkRace(b *testing.B) {
	ctx := context.Background()
	fn := func(_ context.Context) (int, error) { return 1, nil }
	for b.Loop() {
		_, _ = Race(ctx, fn, fn, fn)
	}
}
