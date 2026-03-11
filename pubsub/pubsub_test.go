package pubsub

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestBus_PublishSubscribe(t *testing.T) {
	bus := New[string]()
	var got string
	bus.Subscribe(func(_ context.Context, e string) {
		got = e
	})

	bus.Publish(context.Background(), "hello")
	if got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestBus_MultipleSubscribers(t *testing.T) {
	bus := New[int]()
	var sum atomic.Int64
	for range 3 {
		bus.Subscribe(func(_ context.Context, e int) {
			sum.Add(int64(e))
		})
	}

	bus.Publish(context.Background(), 10)
	if got := sum.Load(); got != 30 {
		t.Errorf("sum = %d, want 30", got)
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := New[int]()
	var count atomic.Int64
	tok := bus.Subscribe(func(_ context.Context, _ int) {
		count.Add(1)
	})

	bus.Publish(context.Background(), 1)
	bus.Unsubscribe(tok)
	bus.Publish(context.Background(), 2)

	if got := count.Load(); got != 1 {
		t.Errorf("count = %d, want 1", got)
	}
}

func TestBus_Unsubscribe_Nonexistent(t *testing.T) {
	bus := New[int]()
	bus.Unsubscribe(Token(99999)) // should not panic
}

func TestBus_PublishAsync(t *testing.T) {
	bus := New[int]()
	var wg sync.WaitGroup
	var sum atomic.Int64

	for range 3 {
		wg.Add(1)
		bus.Subscribe(func(_ context.Context, e int) {
			defer wg.Done()
			sum.Add(int64(e))
		})
	}

	bus.PublishAsync(context.Background(), 5)
	wg.Wait()

	if got := sum.Load(); got != 15 {
		t.Errorf("sum = %d, want 15", got)
	}
}

func TestBus_PublishAsync_Returns_Immediately(t *testing.T) {
	bus := New[int]()
	done := make(chan struct{})
	bus.Subscribe(func(_ context.Context, _ int) {
		<-done
	})

	start := time.Now()
	bus.PublishAsync(context.Background(), 1)
	elapsed := time.Since(start)

	close(done)

	if elapsed > 50*time.Millisecond {
		t.Errorf("PublishAsync took %v, expected < 50ms", elapsed)
	}
}

func TestBus_PanicRecovery(t *testing.T) {
	bus := New[int]()
	var count atomic.Int64

	bus.Subscribe(func(_ context.Context, _ int) {
		panic("test panic")
	})
	bus.Subscribe(func(_ context.Context, _ int) {
		count.Add(1)
	})

	bus.Publish(context.Background(), 1)
	if got := count.Load(); got != 1 {
		t.Errorf("count = %d, want 1 (second handler should still run)", got)
	}
}

func TestBus_Len(t *testing.T) {
	bus := New[int]()
	if bus.Len() != 0 {
		t.Errorf("Len() = %d, want 0", bus.Len())
	}

	tok1 := bus.Subscribe(func(_ context.Context, _ int) {})
	tok2 := bus.Subscribe(func(_ context.Context, _ int) {})
	if bus.Len() != 2 {
		t.Errorf("Len() = %d, want 2", bus.Len())
	}

	bus.Unsubscribe(tok1)
	if bus.Len() != 1 {
		t.Errorf("Len() = %d, want 1", bus.Len())
	}

	bus.Unsubscribe(tok2)
	if bus.Len() != 0 {
		t.Errorf("Len() = %d, want 0", bus.Len())
	}
}

func TestBus_ContextPropagation(t *testing.T) {
	type key struct{}
	bus := New[string]()
	var got string

	bus.Subscribe(func(ctx context.Context, _ string) {
		got, _ = ctx.Value(key{}).(string)
	})

	ctx := context.WithValue(context.Background(), key{}, "from-ctx")
	bus.Publish(ctx, "ignored")

	if got != "from-ctx" {
		t.Errorf("got %q, want %q", got, "from-ctx")
	}
}

func TestBus_NoSubscribers(t *testing.T) {
	bus := New[int]()
	bus.Publish(context.Background(), 42) // should not panic
}

func TestBus_ConcurrentPublishSubscribe(t *testing.T) {
	bus := New[int]()
	var sum atomic.Int64

	var wg sync.WaitGroup
	// Subscribe from multiple goroutines
	for i := range 10 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			bus.Subscribe(func(_ context.Context, e int) {
				sum.Add(int64(e))
			})
			bus.Publish(context.Background(), n)
		}(i)
	}
	wg.Wait()
}

func BenchmarkBus_Publish(b *testing.B) {
	bus := New[int]()
	bus.Subscribe(func(_ context.Context, _ int) {})
	ctx := context.Background()

	for b.Loop() {
		bus.Publish(ctx, 42)
	}
}

func BenchmarkBus_PublishAsync(b *testing.B) {
	bus := New[int]()
	bus.Subscribe(func(_ context.Context, _ int) {})
	ctx := context.Background()

	for b.Loop() {
		bus.PublishAsync(ctx, 42)
	}
}

func BenchmarkBus_Subscribe(b *testing.B) {
	bus := New[int]()
	h := func(_ context.Context, _ int) {}

	for b.Loop() {
		bus.Subscribe(h)
	}
}
