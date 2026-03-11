package types

import "cmp"

// Interval represents a closed range [Lo, Hi] over any ordered type.
type Interval[T cmp.Ordered] struct {
	Lo T
	Hi T
}

// NewInterval creates an interval [lo, hi]. If lo > hi the values are
// swapped so the interval is always valid.
func NewInterval[T cmp.Ordered](lo, hi T) Interval[T] {
	if lo > hi {
		lo, hi = hi, lo
	}
	return Interval[T]{Lo: lo, Hi: hi}
}

// Contains reports whether v falls within [Lo, Hi].
func (iv Interval[T]) Contains(v T) bool {
	return v >= iv.Lo && v <= iv.Hi
}

// Overlaps reports whether iv and other share at least one point.
func (iv Interval[T]) Overlaps(other Interval[T]) bool {
	return iv.Lo <= other.Hi && other.Lo <= iv.Hi
}

// Merge returns the smallest interval containing both iv and other.
// Merge does not require the intervals to overlap.
func (iv Interval[T]) Merge(other Interval[T]) Interval[T] {
	lo := iv.Lo
	if other.Lo < lo {
		lo = other.Lo
	}
	hi := iv.Hi
	if other.Hi > hi {
		hi = other.Hi
	}
	return Interval[T]{Lo: lo, Hi: hi}
}

// Intersect returns the overlapping sub-interval and true, or a zero
// interval and false if there is no overlap.
func (iv Interval[T]) Intersect(other Interval[T]) (Interval[T], bool) {
	if !iv.Overlaps(other) {
		var zero Interval[T]
		return zero, false
	}
	lo := iv.Lo
	if other.Lo > lo {
		lo = other.Lo
	}
	hi := iv.Hi
	if other.Hi < hi {
		hi = other.Hi
	}
	return Interval[T]{Lo: lo, Hi: hi}, true
}

// Empty reports whether the interval contains no elements (Lo > Hi).
// Because NewInterval normalises inputs, this is only possible when
// constructed via a zero-value literal.
func (iv Interval[T]) Empty() bool {
	return iv.Lo > iv.Hi
}

// Equal reports whether iv and other represent the same interval.
func (iv Interval[T]) Equal(other Interval[T]) bool {
	return iv.Lo == other.Lo && iv.Hi == other.Hi
}

// Clamp returns v constrained to [Lo, Hi].
func (iv Interval[T]) Clamp(v T) T {
	if v < iv.Lo {
		return iv.Lo
	}
	if v > iv.Hi {
		return iv.Hi
	}
	return v
}
