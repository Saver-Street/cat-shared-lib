package batch

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
)

func TestChunk_Basic(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4, 5}, 2)
	want := [][]int{{1, 2}, {3, 4}, {5}}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if len(got[i]) != len(want[i]) {
			t.Errorf("chunk[%d] len = %d, want %d", i, len(got[i]), len(want[i]))
		}
	}
}

func TestChunk_ExactDivisor(t *testing.T) {
	got := Chunk([]int{1, 2, 3, 4}, 2)
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestChunk_SingleElement(t *testing.T) {
	got := Chunk([]string{"a"}, 5)
	if len(got) != 1 || len(got[0]) != 1 {
		t.Errorf("got %v, want [[a]]", got)
	}
}

func TestChunk_Empty(t *testing.T) {
	got := Chunk([]int{}, 3)
	if got != nil {
		t.Errorf("got %v, want nil", got)
	}
}

func TestChunk_SizeOne(t *testing.T) {
	got := Chunk([]int{1, 2, 3}, 1)
	if len(got) != 3 {
		t.Fatalf("len = %d, want 3", len(got))
	}
	for i, ch := range got {
		if len(ch) != 1 || ch[0] != i+1 {
			t.Errorf("chunk[%d] = %v, want [%d]", i, ch, i+1)
		}
	}
}

func TestChunk_SizeLargerThanSlice(t *testing.T) {
	got := Chunk([]int{1, 2}, 100)
	if len(got) != 1 || len(got[0]) != 2 {
		t.Errorf("got %v, want [[1 2]]", got)
	}
}

func TestChunk_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for size 0")
		}
	}()
	Chunk([]int{1}, 0)
}

func TestProcess_Sequential(t *testing.T) {
	var processed []int
	items := []int{1, 2, 3, 4, 5}
	err := Process(context.Background(), items, 2, func(_ context.Context, batch []int) error {
		processed = append(processed, batch...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(processed) != 5 {
		t.Errorf("processed %d items, want 5", len(processed))
	}
}

func TestProcess_StopsOnError(t *testing.T) {
	var count int
	items := []int{1, 2, 3, 4, 5, 6}
	err := Process(context.Background(), items, 2, func(_ context.Context, batch []int) error {
		count++
		if count == 2 {
			return errors.New("batch failed")
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if count != 2 {
		t.Errorf("count = %d, want 2 (should stop at error)", count)
	}
}

func TestProcess_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Process(ctx, []int{1, 2, 3}, 1, func(_ context.Context, _ []int) error {
		t.Error("fn should not be called with cancelled context")
		return nil
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestProcess_Empty(t *testing.T) {
	err := Process(context.Background(), []int{}, 2, func(_ context.Context, _ []int) error {
		t.Error("fn should not be called for empty slice")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessConcurrent_AllProcessed(t *testing.T) {
	var count atomic.Int64
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}

	err := ProcessConcurrent(context.Background(), items, 10, 4, func(_ context.Context, batch []int) error {
		count.Add(int64(len(batch)))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 100 {
		t.Errorf("processed %d items, want 100", count.Load())
	}
}

func TestProcessConcurrent_ConcurrencyLimit(t *testing.T) {
	var concurrent, maxConcurrent atomic.Int64
	items := make([]int, 20)

	err := ProcessConcurrent(context.Background(), items, 1, 3, func(_ context.Context, _ []int) error {
		n := concurrent.Add(1)
		for {
			old := maxConcurrent.Load()
			if n <= old || maxConcurrent.CompareAndSwap(old, n) {
				break
			}
		}
		concurrent.Add(-1)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if maxConcurrent.Load() > 3 {
		t.Errorf("max concurrent = %d, should be <= 3", maxConcurrent.Load())
	}
}

func TestProcessConcurrent_CollectsErrors(t *testing.T) {
	items := []int{1, 2, 3, 4}
	err := ProcessConcurrent(context.Background(), items, 1, 2, func(_ context.Context, batch []int) error {
		if batch[0]%2 == 0 {
			return fmt.Errorf("even: %d", batch[0])
		}
		return nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProcessConcurrent_Empty(t *testing.T) {
	err := ProcessConcurrent(context.Background(), []int{}, 2, 4, func(_ context.Context, _ []int) error {
		t.Error("fn should not be called")
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProcessConcurrent_DefaultWorkers(t *testing.T) {
	var count atomic.Int64
	err := ProcessConcurrent(context.Background(), []int{1, 2, 3}, 1, 0, func(_ context.Context, batch []int) error {
		count.Add(int64(len(batch)))
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count.Load() != 3 {
		t.Errorf("processed %d items, want 3", count.Load())
	}
}

func TestMap_Basic(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	results, err := Map(context.Background(), items, 2, func(_ context.Context, item int) (string, error) {
		return fmt.Sprintf("item-%d", item), nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 5 {
		t.Fatalf("len = %d, want 5", len(results))
	}
	if results[0] != "item-1" {
		t.Errorf("results[0] = %q, want %q", results[0], "item-1")
	}
	if results[4] != "item-5" {
		t.Errorf("results[4] = %q, want %q", results[4], "item-5")
	}
}

func TestMap_StopsOnError(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	_, err := Map(context.Background(), items, 2, func(_ context.Context, item int) (int, error) {
		if item == 3 {
			return 0, errors.New("failed at 3")
		}
		return item * 2, nil
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMap_Empty(t *testing.T) {
	results, err := Map(context.Background(), []int{}, 2, func(_ context.Context, item int) (int, error) {
		return item, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("len = %d, want 0", len(results))
	}
}

func TestMap_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Map(ctx, []int{1, 2, 3}, 1, func(_ context.Context, item int) (int, error) {
		return item, nil
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func BenchmarkChunk(b *testing.B) {
	items := make([]int, 1000)
	for b.Loop() {
		Chunk(items, 100)
	}
}

func BenchmarkProcess(b *testing.B) {
	items := make([]int, 1000)
	ctx := context.Background()
	for b.Loop() {
		_ = Process(ctx, items, 100, func(_ context.Context, _ []int) error {
			return nil
		})
	}
}

func BenchmarkProcessConcurrent(b *testing.B) {
	items := make([]int, 1000)
	ctx := context.Background()
	for b.Loop() {
		_ = ProcessConcurrent(ctx, items, 100, 4, func(_ context.Context, _ []int) error {
			return nil
		})
	}
}

func BenchmarkMap(b *testing.B) {
	items := make([]int, 1000)
	ctx := context.Background()
	for b.Loop() {
		_, _ = Map(ctx, items, 100, func(_ context.Context, item int) (int, error) {
			return item * 2, nil
		})
	}
}
