package types

// Result holds either a value of type T or an error, providing a monadic
// approach to error handling.  Use [OK] to create a success result and
// [Fail] to create a failure result.
type Result[T any] struct {
	value T
	err   error
}

// OK creates a successful Result containing the given value.
func OK[T any](v T) Result[T] {
	return Result[T]{value: v}
}

// Fail creates a failed Result containing the given error.
func Fail[T any](err error) Result[T] {
	return Result[T]{err: err}
}

// FromPair creates a Result from a standard Go (value, error) pair.
func FromPair[T any](v T, err error) Result[T] {
	if err != nil {
		return Fail[T](err)
	}
	return OK(v)
}

// IsOK returns true if the result is successful.
func (r Result[T]) IsOK() bool { return r.err == nil }

// IsErr returns true if the result contains an error.
func (r Result[T]) IsErr() bool { return r.err != nil }

// Value returns the contained value.  If the result is an error, the zero
// value of T is returned.
func (r Result[T]) Value() T { return r.value }

// Err returns the contained error, or nil if the result is successful.
func (r Result[T]) Err() error { return r.err }

// Unwrap returns the value and error as a standard Go pair.
func (r Result[T]) Unwrap() (T, error) { return r.value, r.err }

// OrElse returns the contained value if successful, otherwise returns
// the provided fallback.
func (r Result[T]) OrElse(fallback T) T {
	if r.err != nil {
		return fallback
	}
	return r.value
}

// Map transforms the value of a successful result using fn.  If the result
// is a failure, the error is propagated unchanged.
func Map[T, U any](r Result[T], fn func(T) U) Result[U] {
	if r.err != nil {
		return Fail[U](r.err)
	}
	return OK(fn(r.value))
}

// FlatMap transforms the value of a successful result using fn which itself
// returns a Result.  If the original result is a failure, the error is
// propagated unchanged.
func FlatMap[T, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.err != nil {
		return Fail[U](r.err)
	}
	return fn(r.value)
}
