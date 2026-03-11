package types

import (
	"testing"
)

func TestNewRingPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for cap < 1")
		}
		if r != "ring: capacity must be at least 1" {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	NewRing[int](0)
}

func TestRingPushPop(t *testing.T) {
	t.Parallel()
	r := NewRing[int](3)

	if r.Len() != 0 {
		t.Errorf("Len() = %d; want 0", r.Len())
	}
	if r.Cap() != 3 {
		t.Errorf("Cap() = %d; want 3", r.Cap())
	}

	if dropped := r.Push(1); dropped {
		t.Error("Push(1) dropped unexpectedly")
	}
	if dropped := r.Push(2); dropped {
		t.Error("Push(2) dropped unexpectedly")
	}
	if dropped := r.Push(3); dropped {
		t.Error("Push(3) dropped unexpectedly")
	}
	if !r.Full() {
		t.Error("Full() = false; want true")
	}

	v, ok := r.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop() = (%d, %v); want (1, true)", v, ok)
	}
	v, ok = r.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%d, %v); want (2, true)", v, ok)
	}
	v, ok = r.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop() = (%d, %v); want (3, true)", v, ok)
	}
	_, ok = r.Pop()
	if ok {
		t.Error("Pop() on empty ring should return false")
	}
}

func TestRingOverwrite(t *testing.T) {
	t.Parallel()
	r := NewRing[string](2)

	r.Push("a")
	r.Push("b")
	if dropped := r.Push("c"); !dropped {
		t.Error("Push(c) should have dropped oldest")
	}

	got := r.ToSlice()
	want := []string{"b", "c"}
	if len(got) != len(want) {
		t.Fatalf("ToSlice() len = %d; want %d", len(got), len(want))
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("ToSlice()[%d] = %q; want %q", i, v, want[i])
		}
	}
}

func TestRingPeek(t *testing.T) {
	t.Parallel()
	r := NewRing[int](2)

	_, ok := r.Peek()
	if ok {
		t.Error("Peek() on empty ring should return false")
	}

	r.Push(10)
	v, ok := r.Peek()
	if !ok || v != 10 {
		t.Errorf("Peek() = (%d, %v); want (10, true)", v, ok)
	}
	if r.Len() != 1 {
		t.Error("Peek() should not change length")
	}
}

func TestRingClear(t *testing.T) {
	t.Parallel()
	r := NewRing[int](3)
	r.Push(1)
	r.Push(2)
	r.Push(3)
	r.Clear()

	if r.Len() != 0 {
		t.Errorf("Len() after Clear = %d; want 0", r.Len())
	}
	if r.Full() {
		t.Error("Full() after Clear should be false")
	}
	_, ok := r.Pop()
	if ok {
		t.Error("Pop() after Clear should return false")
	}
}

func TestRingDo(t *testing.T) {
	t.Parallel()
	r := NewRing[int](3)
	r.Push(10)
	r.Push(20)
	r.Push(30)

	var got []int
	r.Do(func(v int) { got = append(got, v) })

	want := []int{10, 20, 30}
	if len(got) != len(want) {
		t.Fatalf("Do collected %d; want %d", len(got), len(want))
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("Do[%d] = %d; want %d", i, v, want[i])
		}
	}
}

func TestRingDoAfterOverwrite(t *testing.T) {
	t.Parallel()
	r := NewRing[int](2)
	r.Push(1)
	r.Push(2)
	r.Push(3) // overwrites 1

	var got []int
	r.Do(func(v int) { got = append(got, v) })

	want := []int{2, 3}
	if len(got) != len(want) {
		t.Fatalf("Do collected %d; want %d", len(got), len(want))
	}
	for i, v := range got {
		if v != want[i] {
			t.Errorf("Do[%d] = %d; want %d", i, v, want[i])
		}
	}
}

func TestRingCapOne(t *testing.T) {
	t.Parallel()
	r := NewRing[int](1)

	r.Push(1)
	if r.Len() != 1 {
		t.Errorf("Len() = %d; want 1", r.Len())
	}
	if dropped := r.Push(2); !dropped {
		t.Error("Push(2) on cap-1 ring should drop")
	}
	v, ok := r.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop() = (%d, %v); want (2, true)", v, ok)
	}
}

func TestRingToSliceEmpty(t *testing.T) {
	t.Parallel()
	r := NewRing[int](3)
	s := r.ToSlice()
	if len(s) != 0 {
		t.Errorf("ToSlice() on empty = %v; want []", s)
	}
}

func BenchmarkRingPushPop(b *testing.B) {
	r := NewRing[int](1024)
	for range b.N {
		r.Push(42)
		r.Pop()
	}
}

func BenchmarkRingOverwrite(b *testing.B) {
	r := NewRing[int](64)
	for range b.N {
		r.Push(42)
	}
}

func FuzzRingPush(f *testing.F) {
	f.Add(1, 5)
	f.Add(100, 200)
	f.Fuzz(func(t *testing.T, cap int, n int) {
		if cap < 1 || cap > 10000 {
			return
		}
		if n < 0 || n > 20000 {
			return
		}
		r := NewRing[int](cap)
		for i := range n {
			r.Push(i)
		}
		if r.Len() > r.Cap() {
			t.Errorf("Len %d > Cap %d", r.Len(), r.Cap())
		}
	})
}
