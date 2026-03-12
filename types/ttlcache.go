package types

import (
	"sync"
	"time"
)

type ttlEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// TTLCache is a thread-safe cache where each entry has a time-to-live.
// Expired entries are lazily removed on access.
type TTLCache[K comparable, V any] struct {
	mu         sync.RWMutex
	items      map[K]ttlEntry[V]
	defaultTTL time.Duration
}

// NewTTLCache creates a TTLCache with the given default TTL for entries.
func NewTTLCache[K comparable, V any](defaultTTL time.Duration) *TTLCache[K, V] {
	return &TTLCache[K, V]{
		items:      make(map[K]ttlEntry[V]),
		defaultTTL: defaultTTL,
	}
}

// Set adds or updates a key with the default TTL.
func (c *TTLCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.defaultTTL)
}

// SetWithTTL adds or updates a key with a specific TTL.
func (c *TTLCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = ttlEntry[V]{value: value, expiresAt: time.Now().Add(ttl)}
	c.mu.Unlock()
}

// Get retrieves a value by key. Returns the value and true if found and
// not expired. Expired entries are deleted on access.
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	if time.Now().After(entry.expiresAt) {
		delete(c.items, key)
		var zero V
		return zero, false
	}
	return entry.value, true
}

// Delete removes a key from the cache.
func (c *TTLCache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Len returns the number of entries (including expired but not yet cleaned).
func (c *TTLCache[K, V]) Len() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

// Purge removes all expired entries and returns the number removed.
func (c *TTLCache[K, V]) Purge() int {
	c.mu.Lock()
	now := time.Now()
	removed := 0
	for k, entry := range c.items {
		if now.After(entry.expiresAt) {
			delete(c.items, k)
			removed++
		}
	}
	c.mu.Unlock()
	return removed
}

// Clear removes all entries from the cache.
func (c *TTLCache[K, V]) Clear() {
	c.mu.Lock()
	c.items = make(map[K]ttlEntry[V])
	c.mu.Unlock()
}

// Keys returns all non-expired keys.
func (c *TTLCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now()
	keys := make([]K, 0, len(c.items))
	for k, entry := range c.items {
		if !now.After(entry.expiresAt) {
			keys = append(keys, k)
		}
	}
	return keys
}
