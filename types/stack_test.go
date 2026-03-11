package types

import "testing"

func TestStackNew(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	if s.Len() != 0 {
		t.Errorf("Len = %d; want 0", s.Len())
	}
	if !s.IsEmpty() {
		t.Error("IsEmpty = false; want true")
	}
}

func TestStackZeroValue(t *testing.T) {
	t.Parallel()
	var s Stack[int]
	s.Push(1)
	v, ok := s.Pop()
	if !ok || v != 1 {
		t.Errorf("Pop = %d, %v; want 1, true", v, ok)
	}
}

func TestStackPushPop(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	if s.Len() != 3 {
		t.Fatalf("Len = %d; want 3", s.Len())
	}
	v, ok := s.Pop()
	if !ok || v != 3 {
		t.Errorf("Pop = %d, %v; want 3, true", v, ok)
	}
	v, ok = s.Pop()
	if !ok || v != 2 {
		t.Errorf("Pop = %d, %v; want 2, true", v, ok)
	}
}

func TestStackPopEmpty(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	v, ok := s.Pop()
	if ok || v != 0 {
		t.Errorf("Pop empty = %d, %v; want 0, false", v, ok)
	}
}

func TestStackPeek(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	s.Push(42)
	v, ok := s.Peek()
	if !ok || v != 42 {
		t.Errorf("Peek = %d, %v; want 42, true", v, ok)
	}
	if s.Len() != 1 {
		t.Errorf("Len after Peek = %d; want 1", s.Len())
	}
}

func TestStackPeekEmpty(t *testing.T) {
	t.Parallel()
	s := NewStack[string]()
	v, ok := s.Peek()
	if ok || v != "" {
		t.Errorf("Peek empty = %q, %v; want \"\", false", v, ok)
	}
}

func TestStackClear(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Clear()
	if !s.IsEmpty() {
		t.Error("IsEmpty after Clear = false")
	}
	if s.Len() != 0 {
		t.Errorf("Len after Clear = %d; want 0", s.Len())
	}
}

func TestStackValues(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	s.Push(1)
	s.Push(2)
	s.Push(3)
	vals := s.Values()
	if len(vals) != 3 || vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("Values = %v; want [1 2 3]", vals)
	}
	// Mutating the slice should not affect the stack.
	vals[0] = 99
	v, _ := s.Pop()
	if v != 3 {
		t.Errorf("Pop after Values mutation = %d; want 3", v)
	}
}

func TestStackValuesEmpty(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	vals := s.Values()
	if vals != nil {
		t.Errorf("Values empty = %v; want nil", vals)
	}
}

func TestStackIsEmpty(t *testing.T) {
	t.Parallel()
	s := NewStack[int]()
	if !s.IsEmpty() {
		t.Error("new stack IsEmpty = false")
	}
	s.Push(1)
	if s.IsEmpty() {
		t.Error("after Push IsEmpty = true")
	}
	s.Pop()
	if !s.IsEmpty() {
		t.Error("after Pop IsEmpty = false")
	}
}

func BenchmarkStackPush(b *testing.B) {
	s := NewStack[int]()
	for i := range b.N {
		s.Push(i)
	}
}

func BenchmarkStackPushPop(b *testing.B) {
	s := NewStack[int]()
	for i := range b.N {
		s.Push(i)
		s.Pop()
	}
}

func FuzzStackPushPop(f *testing.F) {
	f.Add(42)
	f.Add(0)
	f.Add(-1)
	f.Fuzz(func(t *testing.T, v int) {
		s := NewStack[int]()
		s.Push(v)
		got, ok := s.Pop()
		if !ok || got != v {
			t.Errorf("Push(%d)/Pop = %d, %v", v, got, ok)
		}
	})
}
