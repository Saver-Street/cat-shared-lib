package types

import "iter"

// Stack is a generic LIFO stack backed by a slice.
type Stack[T any] struct {
	data []T
}

// NewStack creates an empty Stack.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{}
}

// Push adds a value to the top of the stack.
func (s *Stack[T]) Push(v T) {
	s.data = append(s.data, v)
}

// Pop removes and returns the top value. It returns the zero value and
// false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	v := s.data[len(s.data)-1]
	var zero T
	s.data[len(s.data)-1] = zero // clear for GC
	s.data = s.data[:len(s.data)-1]
	return v, true
}

// Peek returns the top value without removing it. It returns the zero
// value and false if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.data) == 0 {
		var zero T
		return zero, false
	}
	return s.data[len(s.data)-1], true
}

// Len returns the number of elements on the stack.
func (s *Stack[T]) Len() int {
	return len(s.data)
}

// IsEmpty reports whether the stack has no elements.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.data) == 0
}

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	s.data = s.data[:0]
}

// All returns an iterator over the stack elements from top to bottom.
func (s *Stack[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := len(s.data) - 1; i >= 0; i-- {
			if !yield(s.data[i]) {
				return
			}
		}
	}
}
