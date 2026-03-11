package cond

import "cmp"

// Ternary returns trueVal if condition is true, falseVal otherwise.
// This is the Go equivalent of the ternary operator (cond ? a : b).
func Ternary[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

// Coalesce returns the first non-zero value from the given values.
// Returns the zero value of T if all values are zero.
func Coalesce[T comparable](vals ...T) T {
	var zero T
	for _, v := range vals {
		if v != zero {
			return v
		}
	}
	return zero
}

// CoalesceFunc returns the result of the first function that returns a
// non-zero value. Functions are evaluated lazily.
func CoalesceFunc[T comparable](fns ...func() T) T {
	var zero T
	for _, fn := range fns {
		if v := fn(); v != zero {
			return v
		}
	}
	return zero
}

// Clamp restricts v to the range [lo, hi].
func Clamp[T cmp.Ordered](v, lo, hi T) T {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Min returns the smallest of the given values. Panics if no values given.
func Min[T cmp.Ordered](vals ...T) T {
	if len(vals) == 0 {
		panic("cond.Min: no values")
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// Max returns the largest of the given values. Panics if no values given.
func Max[T cmp.Ordered](vals ...T) T {
	if len(vals) == 0 {
		panic("cond.Max: no values")
	}
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Zero returns the zero value of type T.
func Zero[T any]() T {
	var zero T
	return zero
}

// IsZero returns true if v equals the zero value of its type.
func IsZero[T comparable](v T) bool {
	var zero T
	return v == zero
}

// Case represents a single condition-result pair for use with [Switch].
type Case[T any] struct {
	When bool
	Then T
}

// Switch returns the Then value of the first Case where When is true.
// Returns the zero value of T if no case matches.
func Switch[T any](cases ...Case[T]) T {
	for _, c := range cases {
		if c.When {
			return c.Then
		}
	}
	var zero T
	return zero
}
