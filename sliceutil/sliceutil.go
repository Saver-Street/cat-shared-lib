package sliceutil

// Partition splits items into two slices: one where predicate returns true,
// one where it returns false.
func Partition[T any](items []T, predicate func(T) bool) (matched, unmatched []T) {
	matched = make([]T, 0)
	unmatched = make([]T, 0)
	for _, v := range items {
		if predicate(v) {
			matched = append(matched, v)
		} else {
			unmatched = append(unmatched, v)
		}
	}
	return matched, unmatched
}

// Reduce applies fn to each element with an accumulator, returning the
// final accumulated value.
func Reduce[T any, U any](items []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range items {
		acc = fn(acc, v)
	}
	return acc
}

// Any returns true if predicate returns true for any element.
func Any[T any](items []T, predicate func(T) bool) bool {
	for _, v := range items {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All returns true if predicate returns true for all elements (or slice is empty).
func All[T any](items []T, predicate func(T) bool) bool {
	for _, v := range items {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None returns true if predicate returns false for all elements.
func None[T any](items []T, predicate func(T) bool) bool {
	for _, v := range items {
		if predicate(v) {
			return false
		}
	}
	return true
}

// Find returns the first element matching predicate and true, or zero and false.
func Find[T any](items []T, predicate func(T) bool) (T, bool) {
	for _, v := range items {
		if predicate(v) {
			return v, true
		}
	}
	var zero T
	return zero, false
}

// FindIndex returns the index of the first element matching predicate, or -1.
func FindIndex[T any](items []T, predicate func(T) bool) int {
	for i, v := range items {
		if predicate(v) {
			return i
		}
	}
	return -1
}

// Last returns the last element of the slice, or zero and false if empty.
func Last[T any](items []T) (T, bool) {
	if len(items) == 0 {
		var zero T
		return zero, false
	}
	return items[len(items)-1], true
}

// Take returns the first n elements. If n > len, returns all elements.
func Take[T any](items []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(items) {
		n = len(items)
	}
	return items[:n]
}

// Drop returns all elements after the first n. If n > len, returns empty.
func Drop[T any](items []T, n int) []T {
	if n < 0 {
		n = 0
	}
	if n > len(items) {
		n = len(items)
	}
	return items[n:]
}

// Associate creates a map from a slice using a key function.
func Associate[T any, K comparable](items []T, keyFn func(T) K) map[K]T {
	m := make(map[K]T, len(items))
	for _, v := range items {
		m[keyFn(v)] = v
	}
	return m
}

// FlatMap applies fn to each element and flattens the results.
func FlatMap[T any, U any](items []T, fn func(T) []U) []U {
	var result []U
	for _, v := range items {
		result = append(result, fn(v)...)
	}
	return result
}

// Count returns the number of elements matching predicate.
func Count[T any](items []T, predicate func(T) bool) int {
	n := 0
	for _, v := range items {
		if predicate(v) {
			n++
		}
	}
	return n
}

// ForEach calls fn for each element with its index.
func ForEach[T any](items []T, fn func(int, T)) {
	for i, v := range items {
		fn(i, v)
	}
}
