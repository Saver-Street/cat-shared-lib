package types

import (
	"testing"
)

func TestPriorityQueueBasic(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	if !pq.IsEmpty() {
		t.Error("new queue should be empty")
	}
	if pq.Len() != 0 {
		t.Errorf("Len() = %d; want 0", pq.Len())
	}

	pq.Push("low", 1)
	pq.Push("high", 10)
	pq.Push("mid", 5)

	if pq.Len() != 3 {
		t.Errorf("Len() = %d; want 3", pq.Len())
	}
	if pq.IsEmpty() {
		t.Error("queue should not be empty")
	}
}

func TestPriorityQueueMinOrder(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	pq.Push("c", 3)
	pq.Push("a", 1)
	pq.Push("b", 2)

	want := []string{"a", "b", "c"}
	for i, w := range want {
		got := pq.Pop()
		if got != w {
			t.Errorf("Pop()[%d] = %q; want %q", i, got, w)
		}
	}
}

func TestPriorityQueueMaxOrder(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	// Negate priorities for max-heap behavior.
	pq.Push("a", -1)
	pq.Push("c", -3)
	pq.Push("b", -2)

	want := []string{"c", "b", "a"}
	for i, w := range want {
		got := pq.Pop()
		if got != w {
			t.Errorf("Pop()[%d] = %q; want %q", i, got, w)
		}
	}
}

func TestPriorityQueuePeek(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[int]()

	pq.Push(42, 2)
	pq.Push(7, 1)

	if got := pq.Peek(); got != 7 {
		t.Errorf("Peek() = %d; want 7", got)
	}
	// Peek should not remove
	if pq.Len() != 2 {
		t.Errorf("Len() = %d after Peek; want 2", pq.Len())
	}
}

func TestPriorityQueueDuplicatePriorities(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	pq.Push("first", 1)
	pq.Push("second", 1)
	pq.Push("third", 1)

	if pq.Len() != 3 {
		t.Errorf("Len() = %d; want 3", pq.Len())
	}
	// All items should be dequeued
	for range 3 {
		pq.Pop()
	}
	if !pq.IsEmpty() {
		t.Error("queue should be empty after popping all items")
	}
}

func TestPriorityQueueSingleItem(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	pq.Push("only", 5)
	if got := pq.Peek(); got != "only" {
		t.Errorf("Peek() = %q; want only", got)
	}
	if got := pq.Pop(); got != "only" {
		t.Errorf("Pop() = %q; want only", got)
	}
	if !pq.IsEmpty() {
		t.Error("queue should be empty")
	}
}

func TestPriorityQueueManyItems(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[int]()
	n := 1000

	// Push in reverse order
	for i := n; i > 0; i-- {
		pq.Push(i, i)
	}

	// Should come out in ascending order
	for i := 1; i <= n; i++ {
		got := pq.Pop()
		if got != i {
			t.Errorf("Pop() = %d; want %d", got, i)
		}
	}
}

func TestPriorityQueueInterleavedOps(t *testing.T) {
	t.Parallel()
	pq := NewPriorityQueue[string]()

	pq.Push("b", 2)
	pq.Push("d", 4)
	if got := pq.Pop(); got != "b" {
		t.Errorf("Pop() = %q; want b", got)
	}

	pq.Push("a", 1)
	pq.Push("c", 3)
	if got := pq.Pop(); got != "a" {
		t.Errorf("Pop() = %q; want a", got)
	}
	if got := pq.Pop(); got != "c" {
		t.Errorf("Pop() = %q; want c", got)
	}
	if got := pq.Pop(); got != "d" {
		t.Errorf("Pop() = %q; want d", got)
	}
}

func BenchmarkPriorityQueuePush(b *testing.B) {
	pq := NewPriorityQueue[int]()
	for i := range b.N {
		pq.Push(i, i)
	}
}

func BenchmarkPriorityQueuePop(b *testing.B) {
	pq := NewPriorityQueue[int]()
	for i := range b.N {
		pq.Push(i, i)
	}
	b.ResetTimer()
	for range b.N {
		pq.Pop()
	}
}

func BenchmarkPriorityQueuePushPop(b *testing.B) {
	pq := NewPriorityQueue[int]()
	for i := range b.N {
		pq.Push(i, i)
		if i%2 == 0 {
			pq.Pop()
		}
	}
}

func FuzzPriorityQueue(f *testing.F) {
	f.Add(1, 2, 3)
	f.Add(100, 50, 75)
	f.Add(-1, 0, 1)
	f.Fuzz(func(t *testing.T, a, b, c int) {
		pq := NewPriorityQueue[int]()
		pq.Push(a, a)
		pq.Push(b, b)
		pq.Push(c, c)

		prev := pq.Pop()
		for !pq.IsEmpty() {
			curr := pq.Pop()
			if curr < prev {
				t.Errorf("items not in order: %d after %d", curr, prev)
			}
			prev = curr
		}
	})
}
