package types

// Chunk splits a slice into sub-slices of the given size.
// The last chunk may be shorter if the input length is not evenly divisible.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 || len(items) == 0 {
		return nil
	}
	n := (len(items) + size - 1) / size
	chunks := make([][]T, 0, n)
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Flatten concatenates a slice of slices into a single slice.
func Flatten[T any](lists [][]T) []T {
	total := 0
	for _, l := range lists {
		total += len(l)
	}
	out := make([]T, 0, total)
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// GroupBy groups slice elements by a key function.  The returned map keys
// are the result of keyFn applied to each element.
func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	m := make(map[K][]T)
	for _, item := range items {
		k := keyFn(item)
		m[k] = append(m[k], item)
	}
	return m
}

// Reverse returns a new slice with elements in reverse order.
func Reverse[T any](items []T) []T {
	out := make([]T, len(items))
	for i, v := range items {
		out[len(items)-1-i] = v
	}
	return out
}

// Zip combines two slices into a slice of Pair values.  The result length
// equals the shorter input.
func Zip[T, U any](a []T, b []U) []Pair[T, U] {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	out := make([]Pair[T, U], n)
	for i := range n {
		out[i] = Pair[T, U]{First: a[i], Second: b[i]}
	}
	return out
}

// Pair holds two values of different types.
type Pair[T, U any] struct {
	First  T
	Second U
}

// Index returns the first index of target in items, or -1 if not found.
func Index[T comparable](items []T, target T) int {
	for i, v := range items {
		if v == target {
			return i
		}
	}
	return -1
}
