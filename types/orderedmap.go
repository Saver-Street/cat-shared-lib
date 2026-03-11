package types

import "iter"

// OrderedMap is a map that preserves insertion order of keys.
// It provides O(1) lookups and O(1) amortised insertions.
type OrderedMap[K comparable, V any] struct {
	keys   []K
	values map[K]V
}

// NewOrderedMap creates an empty OrderedMap.
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{
		values: make(map[K]V),
	}
}

// Set adds or updates a key-value pair. New keys are appended;
// existing keys retain their original position.
func (m *OrderedMap[K, V]) Set(key K, value V) {
	if _, ok := m.values[key]; !ok {
		m.keys = append(m.keys, key)
	}
	m.values[key] = value
}

// Get returns the value for key and a boolean indicating presence.
func (m *OrderedMap[K, V]) Get(key K) (V, bool) {
	v, ok := m.values[key]
	return v, ok
}

// Delete removes a key and its value. It preserves the order of
// remaining keys.
func (m *OrderedMap[K, V]) Delete(key K) {
	if _, ok := m.values[key]; !ok {
		return
	}
	delete(m.values, key)
	for i, k := range m.keys {
		if k == key {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			return
		}
	}
}

// Len returns the number of entries.
func (m *OrderedMap[K, V]) Len() int {
	return len(m.values)
}

// Has reports whether key exists in the map.
func (m *OrderedMap[K, V]) Has(key K) bool {
	_, ok := m.values[key]
	return ok
}

// Keys returns a copy of the keys in insertion order.
func (m *OrderedMap[K, V]) Keys() []K {
	out := make([]K, len(m.keys))
	copy(out, m.keys)
	return out
}

// Values returns the values in insertion order.
func (m *OrderedMap[K, V]) Values() []V {
	out := make([]V, len(m.keys))
	for i, k := range m.keys {
		out[i] = m.values[k]
	}
	return out
}

// All returns an iterator over key-value pairs in insertion order.
func (m *OrderedMap[K, V]) All() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, k := range m.keys {
			if !yield(k, m.values[k]) {
				return
			}
		}
	}
}
