package types

// Stack is a generic LIFO stack backed by a slice.
// The zero value is an empty stack ready to use.
type Stack[T any] struct {
	items []T
}

// NewStack returns a new empty Stack.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(v T) {
	s.items = append(s.items, v)
}

// Pop removes and returns the top element. The bool is false if the
// stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	v := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return v, true
}

// Peek returns the top element without removing it. The bool is false
// if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// IsEmpty reports whether the stack has no elements.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	s.items = s.items[:0]
}

// Values returns a copy of the elements in bottom-to-top order.
func (s *Stack[T]) Values() []T {
	if len(s.items) == 0 {
		return nil
	}
	cp := make([]T, len(s.items))
	copy(cp, s.items)
	return cp
}
