package types

// Ring is a fixed-capacity circular buffer. When full, new writes
// overwrite the oldest elements.
type Ring[T any] struct {
	buf  []T
	head int
	tail int
	len  int
	cap  int
}

// NewRing creates a ring buffer with the given capacity.
// It panics if cap is less than 1.
func NewRing[T any](cap int) *Ring[T] {
	if cap < 1 {
		panic("ring: capacity must be at least 1")
	}
	return &Ring[T]{
		buf: make([]T, cap),
		cap: cap,
	}
}

// Push adds a value to the ring. If the ring is full the oldest
// element is overwritten and dropped is true.
func (r *Ring[T]) Push(v T) (dropped bool) {
	if r.len == r.cap {
		r.buf[r.tail] = v
		r.tail = (r.tail + 1) % r.cap
		r.head = (r.head + 1) % r.cap
		return true
	}
	r.buf[r.tail] = v
	r.tail = (r.tail + 1) % r.cap
	r.len++
	return false
}

// Pop removes and returns the oldest element. If the ring is empty
// ok is false.
func (r *Ring[T]) Pop() (v T, ok bool) {
	if r.len == 0 {
		var zero T
		return zero, false
	}
	v = r.buf[r.head]
	var zero T
	r.buf[r.head] = zero
	r.head = (r.head + 1) % r.cap
	r.len--
	return v, true
}

// Peek returns the oldest element without removing it.
func (r *Ring[T]) Peek() (v T, ok bool) {
	if r.len == 0 {
		var zero T
		return zero, false
	}
	return r.buf[r.head], true
}

// Len returns the number of elements currently stored.
func (r *Ring[T]) Len() int { return r.len }

// Cap returns the fixed capacity.
func (r *Ring[T]) Cap() int { return r.cap }

// Full reports whether the ring is at capacity.
func (r *Ring[T]) Full() bool { return r.len == r.cap }

// Clear removes all elements.
func (r *Ring[T]) Clear() {
	var zero T
	for i := range r.buf {
		r.buf[i] = zero
	}
	r.head = 0
	r.tail = 0
	r.len = 0
}

// Do calls fn for each element from oldest to newest.
func (r *Ring[T]) Do(fn func(T)) {
	for i := range r.len {
		fn(r.buf[(r.head+i)%r.cap])
	}
}

// ToSlice returns a copy of the elements from oldest to newest.
func (r *Ring[T]) ToSlice() []T {
	s := make([]T, r.len)
	for i := range r.len {
		s[i] = r.buf[(r.head+i)%r.cap]
	}
	return s
}
