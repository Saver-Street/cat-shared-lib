package types

// Set is a generic unordered collection of unique comparable values backed by a
// map.  The zero value is ready to use after calling [NewSet].
type Set[T comparable] struct {
	m map[T]struct{}
}

// NewSet returns a Set initialised with the given values.
func NewSet[T comparable](vals ...T) Set[T] {
	s := Set[T]{m: make(map[T]struct{}, len(vals))}
	for _, v := range vals {
		s.m[v] = struct{}{}
	}
	return s
}

// Add inserts one or more values into the set.
func (s *Set[T]) Add(vals ...T) {
	if s.m == nil {
		s.m = make(map[T]struct{}, len(vals))
	}
	for _, v := range vals {
		s.m[v] = struct{}{}
	}
}

// Remove deletes one or more values from the set.
func (s *Set[T]) Remove(vals ...T) {
	for _, v := range vals {
		delete(s.m, v)
	}
}

// Contains reports whether v is in the set.
func (s Set[T]) Contains(v T) bool {
	_, ok := s.m[v]
	return ok
}

// Len returns the number of elements in the set.
func (s Set[T]) Len() int {
	return len(s.m)
}

// Values returns all elements as a slice in no particular order.
func (s Set[T]) Values() []T {
	out := make([]T, 0, len(s.m))
	for v := range s.m {
		out = append(out, v)
	}
	return out
}

// Union returns a new set containing all elements from both sets.
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := NewSet[T]()
	for v := range s.m {
		result.m[v] = struct{}{}
	}
	for v := range other.m {
		result.m[v] = struct{}{}
	}
	return result
}

// Intersect returns a new set containing only elements present in both sets.
func (s Set[T]) Intersect(other Set[T]) Set[T] {
	result := NewSet[T]()
	// Iterate over the smaller set for efficiency.
	a, b := s, other
	if a.Len() > b.Len() {
		a, b = b, a
	}
	for v := range a.m {
		if b.Contains(v) {
			result.m[v] = struct{}{}
		}
	}
	return result
}

// Diff returns a new set containing elements in s that are not in other.
func (s Set[T]) Diff(other Set[T]) Set[T] {
	result := NewSet[T]()
	for v := range s.m {
		if !other.Contains(v) {
			result.m[v] = struct{}{}
		}
	}
	return result
}

// Equal reports whether s and other contain the same elements.
func (s Set[T]) Equal(other Set[T]) bool {
	if s.Len() != other.Len() {
		return false
	}
	for v := range s.m {
		if !other.Contains(v) {
			return false
		}
	}
	return true
}
