package types

import "sync"

// Lazy[T] is a thread-safe wrapper for deferred computation. The initializer
// function is called at most once, on the first call to Get.
type Lazy[T any] struct {
	once sync.Once
	init func() T
	val  T
}

// NewLazy creates a Lazy value that will call init on first access.
func NewLazy[T any](init func() T) *Lazy[T] {
	return &Lazy[T]{init: init}
}

// Get returns the lazily-initialized value. The init function is called at
// most once, even under concurrent access.
func (l *Lazy[T]) Get() T {
	l.once.Do(func() {
		l.val = l.init()
	})
	return l.val
}

// LazyErr[T] is like Lazy[T] but the initializer may return an error.
// If init returns an error, subsequent calls to Get will retry until init
// succeeds.
type LazyErr[T any] struct {
	mu   sync.Mutex
	init func() (T, error)
	val  T
	done bool
}

// NewLazyErr creates a LazyErr value.
func NewLazyErr[T any](init func() (T, error)) *LazyErr[T] {
	return &LazyErr[T]{init: init}
}

// Get returns the lazily-initialized value. If the init function has not
// succeeded yet, it is called again. Once it succeeds, the result is cached.
func (l *LazyErr[T]) Get() (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.done {
		return l.val, nil
	}
	v, err := l.init()
	if err != nil {
		return v, err
	}
	l.val = v
	l.done = true
	return v, nil
}
