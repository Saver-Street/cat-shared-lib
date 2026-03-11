package types

import "sync"

// SyncMap is a type-safe concurrent map. It wraps sync.RWMutex for safe
// concurrent reads and writes. For highly contended workloads with many
// goroutines, consider using sync.Map instead.
type SyncMap[K comparable, V any] struct {
	mu sync.RWMutex
	m  map[K]V
}

// NewSyncMap returns an initialized SyncMap.
func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{m: make(map[K]V)}
}

// Set stores or updates the value for a key.
func (sm *SyncMap[K, V]) Set(key K, value V) {
	sm.mu.Lock()
	sm.m[key] = value
	sm.mu.Unlock()
}

// Get retrieves the value for a key. The second return value reports whether
// the key was found.
func (sm *SyncMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	v, ok := sm.m[key]
	sm.mu.RUnlock()
	return v, ok
}

// Delete removes a key from the map.
func (sm *SyncMap[K, V]) Delete(key K) {
	sm.mu.Lock()
	delete(sm.m, key)
	sm.mu.Unlock()
}

// Has reports whether the map contains the given key.
func (sm *SyncMap[K, V]) Has(key K) bool {
	sm.mu.RLock()
	_, ok := sm.m[key]
	sm.mu.RUnlock()
	return ok
}

// Len returns the number of entries in the map.
func (sm *SyncMap[K, V]) Len() int {
	sm.mu.RLock()
	n := len(sm.m)
	sm.mu.RUnlock()
	return n
}

// Keys returns all keys in the map. Order is not guaranteed.
func (sm *SyncMap[K, V]) Keys() []K {
	sm.mu.RLock()
	keys := make([]K, 0, len(sm.m))
	for k := range sm.m {
		keys = append(keys, k)
	}
	sm.mu.RUnlock()
	return keys
}

// Values returns all values in the map. Order is not guaranteed.
func (sm *SyncMap[K, V]) Values() []V {
	sm.mu.RLock()
	vals := make([]V, 0, len(sm.m))
	for _, v := range sm.m {
		vals = append(vals, v)
	}
	sm.mu.RUnlock()
	return vals
}

// Range calls fn for each key-value pair. If fn returns false, iteration stops.
func (sm *SyncMap[K, V]) Range(fn func(key K, value V) bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	for k, v := range sm.m {
		if !fn(k, v) {
			return
		}
	}
}

// GetOrSet returns the existing value for a key if present. Otherwise it
// stores and returns the given value. The second return value reports whether
// the value was loaded (true) or stored (false).
func (sm *SyncMap[K, V]) GetOrSet(key K, value V) (V, bool) {
	sm.mu.Lock()
	if v, ok := sm.m[key]; ok {
		sm.mu.Unlock()
		return v, true
	}
	sm.m[key] = value
	sm.mu.Unlock()
	return value, false
}

// Clear removes all entries from the map.
func (sm *SyncMap[K, V]) Clear() {
	sm.mu.Lock()
	sm.m = make(map[K]V)
	sm.mu.Unlock()
}
