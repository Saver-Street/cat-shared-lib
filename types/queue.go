package types

import "iter"

// Queue is a generic FIFO queue backed by a ring buffer that grows as needed.
type Queue[T any] struct {
	buf        []T
	head, tail int
	count      int
}

// NewQueue creates an empty Queue.
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{buf: make([]T, 4)}
}

// Enqueue adds a value to the back of the queue.
func (q *Queue[T]) Enqueue(v T) {
	if q.count == len(q.buf) {
		q.grow()
	}
	q.buf[q.tail] = v
	q.tail = (q.tail + 1) % len(q.buf)
	q.count++
}

// Dequeue removes and returns the front value. It returns the zero value
// and false if the queue is empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	v := q.buf[q.head]
	var zero T
	q.buf[q.head] = zero // clear reference for GC
	q.head = (q.head + 1) % len(q.buf)
	q.count--
	return v, true
}

// Peek returns the front value without removing it. It returns the zero
// value and false if the queue is empty.
func (q *Queue[T]) Peek() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.buf[q.head], true
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int {
	return q.count
}

// IsEmpty reports whether the queue has no elements.
func (q *Queue[T]) IsEmpty() bool {
	return q.count == 0
}

// Clear removes all elements from the queue.
func (q *Queue[T]) Clear() {
	q.buf = make([]T, 4)
	q.head = 0
	q.tail = 0
	q.count = 0
}

// All returns an iterator over the queue elements from front to back
// without modifying the queue.
func (q *Queue[T]) All() iter.Seq[T] {
	return func(yield func(T) bool) {
		for i := range q.count {
			idx := (q.head + i) % len(q.buf)
			if !yield(q.buf[idx]) {
				return
			}
		}
	}
}

func (q *Queue[T]) grow() {
	newBuf := make([]T, len(q.buf)*2)
	for i := range q.count {
		newBuf[i] = q.buf[(q.head+i)%len(q.buf)]
	}
	q.buf = newBuf
	q.head = 0
	q.tail = q.count
}
