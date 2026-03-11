package types

import "testing"

func TestQueueEnqueueDequeue(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	v, ok := q.Dequeue()
	if !ok || v != 1 {
		t.Fatalf("Dequeue() = %d, %v; want 1, true", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 2 {
		t.Fatalf("Dequeue() = %d, %v; want 2, true", v, ok)
	}
	v, ok = q.Dequeue()
	if !ok || v != 3 {
		t.Fatalf("Dequeue() = %d, %v; want 3, true", v, ok)
	}
}

func TestQueueDequeueEmpty(t *testing.T) {
	q := NewQueue[string]()
	v, ok := q.Dequeue()
	if ok {
		t.Fatalf("Dequeue() = %q, true; want zero, false", v)
	}
}

func TestQueuePeek(t *testing.T) {
	q := NewQueue[int]()
	_, ok := q.Peek()
	if ok {
		t.Fatal("Peek() on empty should return false")
	}
	q.Enqueue(42)
	v, ok := q.Peek()
	if !ok || v != 42 {
		t.Fatalf("Peek() = %d, %v; want 42, true", v, ok)
	}
	if q.Len() != 1 {
		t.Fatalf("Peek should not remove element, Len() = %d", q.Len())
	}
}

func TestQueueLen(t *testing.T) {
	q := NewQueue[int]()
	if q.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", q.Len())
	}
	q.Enqueue(1)
	q.Enqueue(2)
	if q.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", q.Len())
	}
	q.Dequeue()
	if q.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", q.Len())
	}
}

func TestQueueIsEmpty(t *testing.T) {
	q := NewQueue[int]()
	if !q.IsEmpty() {
		t.Fatal("IsEmpty() = false, want true")
	}
	q.Enqueue(1)
	if q.IsEmpty() {
		t.Fatal("IsEmpty() = true, want false")
	}
}

func TestQueueClear(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Clear()
	if q.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", q.Len())
	}
	if !q.IsEmpty() {
		t.Fatal("IsEmpty() after Clear = false")
	}
}

func TestQueueGrow(t *testing.T) {
	q := NewQueue[int]()
	// Fill past initial capacity (4) to trigger grow
	for i := range 10 {
		q.Enqueue(i)
	}
	if q.Len() != 10 {
		t.Fatalf("Len() = %d, want 10", q.Len())
	}
	for i := range 10 {
		v, ok := q.Dequeue()
		if !ok || v != i {
			t.Fatalf("Dequeue() = %d, %v; want %d, true", v, ok, i)
		}
	}
}

func TestQueueWrapAround(t *testing.T) {
	q := NewQueue[int]()
	// Enqueue and dequeue to advance head pointer
	for i := range 3 {
		q.Enqueue(i)
	}
	for range 3 {
		q.Dequeue()
	}
	// Now head is at index 3, enqueue wraps around
	q.Enqueue(10)
	q.Enqueue(11)
	v, _ := q.Dequeue()
	if v != 10 {
		t.Fatalf("got %d, want 10", v)
	}
	v, _ = q.Dequeue()
	if v != 11 {
		t.Fatalf("got %d, want 11", v)
	}
}

func TestQueueAll(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	got := make([]int, 0, 3)
	for v := range q.All() {
		got = append(got, v)
	}
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("All() = %v, want [1 2 3]", got)
	}
	// Queue should not be modified
	if q.Len() != 3 {
		t.Fatalf("Len() after All = %d, want 3", q.Len())
	}
}

func TestQueueAllEarlyBreak(t *testing.T) {
	q := NewQueue[int]()
	q.Enqueue(1)
	q.Enqueue(2)
	q.Enqueue(3)

	count := 0
	for range q.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestQueueAllEmpty(t *testing.T) {
	q := NewQueue[int]()
	count := 0
	for range q.All() {
		count++
	}
	if count != 0 {
		t.Fatalf("All() on empty yielded %d items", count)
	}
}

func BenchmarkQueueEnqueue(b *testing.B) {
	for b.Loop() {
		q := NewQueue[int]()
		for i := range 100 {
			q.Enqueue(i)
		}
	}
}

func BenchmarkQueueDequeue(b *testing.B) {
	q := NewQueue[int]()
	for i := range 100 {
		q.Enqueue(i)
	}
	b.ResetTimer()
	for b.Loop() {
		// Re-fill
		for i := range 100 {
			q.Enqueue(i)
		}
		for range 100 {
			q.Dequeue()
		}
	}
}

func FuzzQueueRoundTrip(f *testing.F) {
	f.Add(1, 2, 3)
	f.Add(0, 0, 0)
	f.Add(-1, 100, 42)

	f.Fuzz(func(t *testing.T, a, b, c int) {
		q := NewQueue[int]()
		q.Enqueue(a)
		q.Enqueue(b)
		q.Enqueue(c)
		v1, _ := q.Dequeue()
		v2, _ := q.Dequeue()
		v3, _ := q.Dequeue()
		if v1 != a || v2 != b || v3 != c {
			t.Fatalf("FIFO violation: got %d,%d,%d want %d,%d,%d", v1, v2, v3, a, b, c)
		}
	})
}
