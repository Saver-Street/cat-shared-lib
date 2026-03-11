package types

import "sync"

// Pool[T] is a type-safe object pool. It wraps sync.Pool to provide
// generic, allocation-free reuse of objects.
type Pool[T any] struct {
	p sync.Pool
}

// NewPool creates a pool that uses newFunc to create new instances when
// the pool is empty.
func NewPool[T any](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		p: sync.Pool{
			New: func() any { return newFunc() },
		},
	}
}

// Get retrieves an item from the pool or creates a new one.
func (p *Pool[T]) Get() T {
	return p.p.Get().(T)
}

// Put returns an item to the pool for reuse.
func (p *Pool[T]) Put(v T) {
	p.p.Put(v)
}
