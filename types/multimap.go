package types

// MultiMap is a map from a single key to multiple values.
// The zero value is ready to use.
type MultiMap[K comparable, V any] struct {
	m map[K][]V
}

// NewMultiMap returns an initialised MultiMap.
func NewMultiMap[K comparable, V any]() *MultiMap[K, V] {
	return &MultiMap[K, V]{m: make(map[K][]V)}
}

// Put appends one or more values under the given key.
func (mm *MultiMap[K, V]) Put(key K, values ...V) {
	if mm.m == nil {
		mm.m = make(map[K][]V)
	}
	mm.m[key] = append(mm.m[key], values...)
}

// Get returns the values associated with key and whether the key exists.
func (mm *MultiMap[K, V]) Get(key K) ([]V, bool) {
	if mm.m == nil {
		return nil, false
	}
	v, ok := mm.m[key]
	return v, ok
}

// Delete removes all values for the given key.
func (mm *MultiMap[K, V]) Delete(key K) {
	if mm.m != nil {
		delete(mm.m, key)
	}
}

// Contains reports whether key is present in the map.
func (mm *MultiMap[K, V]) Contains(key K) bool {
	if mm.m == nil {
		return false
	}
	_, ok := mm.m[key]
	return ok
}

// Len returns the number of distinct keys.
func (mm *MultiMap[K, V]) Len() int {
	return len(mm.m)
}

// ValueCount returns the total number of values across all keys.
func (mm *MultiMap[K, V]) ValueCount() int {
	n := 0
	for _, v := range mm.m {
		n += len(v)
	}
	return n
}

// Keys returns all distinct keys in the map. Order is not guaranteed.
func (mm *MultiMap[K, V]) Keys() []K {
	if mm.m == nil {
		return nil
	}
	keys := make([]K, 0, len(mm.m))
	for k := range mm.m {
		keys = append(keys, k)
	}
	return keys
}

// Each calls fn for every key-value pair. Iteration stops early if fn
// returns false.
func (mm *MultiMap[K, V]) Each(fn func(K, V) bool) {
	for k, vals := range mm.m {
		for _, v := range vals {
			if !fn(k, v) {
				return
			}
		}
	}
}

// Clear removes all entries.
func (mm *MultiMap[K, V]) Clear() {
	mm.m = nil
}
