package types

import "testing"

func TestStackPushPop(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Fatalf("Pop() = %d, %v; want 3, true", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Fatalf("Pop() = %d, %v; want 2, true", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 1 {
		t.Fatalf("Pop() = %d, %v; want 1, true", v, ok)
	}
}

func TestStackPopEmpty(t *testing.T) {
	s := NewStack[string]()
	v, ok := s.Pop()
	if ok {
		t.Fatalf("Pop() = %q, true; want zero, false", v)
	}
}

func TestStackPeek(t *testing.T) {
	s := NewStack[int]()
	_, ok := s.Peek()
	if ok {
		t.Fatal("Peek() on empty should return false")
	}
	s.Push(42)
	v, ok := s.Peek()
	if !ok || v != 42 {
		t.Fatalf("Peek() = %d, %v; want 42, true", v, ok)
	}
	if s.Len() != 1 {
		t.Fatal("Peek should not remove element")
	}
}

func TestStackLen(t *testing.T) {
	s := NewStack[int]()
	if s.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", s.Len())
	}
	s.Push(1)
	s.Push(2)
	if s.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", s.Len())
	}
}

func TestStackIsEmpty(t *testing.T) {
	s := NewStack[int]()
	if !s.IsEmpty() {
		t.Fatal("IsEmpty() = false, want true")
	}
	s.Push(1)
	if s.IsEmpty() {
		t.Fatal("IsEmpty() = true, want false")
	}
}

func TestStackClear(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", s.Len())
	}
}

func TestStackAll(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	got := make([]int, 0, 3)
	for v := range s.All() {
		got = append(got, v)
	}
	// Top to bottom: 3, 2, 1
	want := []int{3, 2, 1}
	if len(got) != len(want) {
		t.Fatalf("All() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("All()[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestStackAllEarlyBreak(t *testing.T) {
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)

	count := 0
	for range s.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestStackAllEmpty(t *testing.T) {
	s := NewStack[int]()
	count := 0
	for range s.All() {
		count++
	}
	if count != 0 {
		t.Fatalf("All() on empty yielded %d items", count)
	}
}

func BenchmarkStackPush(b *testing.B) {
	for b.Loop() {
		s := NewStack[int]()
		for i := range 100 {
			s.Push(i)
		}
	}
}

func BenchmarkStackPop(b *testing.B) {
	for b.Loop() {
		s := NewStack[int]()
		for i := range 100 {
			s.Push(i)
		}
		for range 100 {
			s.Pop()
		}
	}
}

func FuzzStackRoundTrip(f *testing.F) {
	f.Add(1, 2, 3)
	f.Add(0, 0, 0)
	f.Add(-1, 100, 42)

	f.Fuzz(func(t *testing.T, a, b, c int) {
		s := NewStack[int]()
		s.Push(a)
		s.Push(b)
		s.Push(c)
		v3, _ := s.Pop()
		v2, _ := s.Pop()
		v1, _ := s.Pop()
		if v1 != a || v2 != b || v3 != c {
			t.Fatalf("LIFO violation: got %d,%d,%d want %d,%d,%d", v1, v2, v3, a, b, c)
		}
	})
}
