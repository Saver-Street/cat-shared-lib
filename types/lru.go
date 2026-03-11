package types

import "container/list"

// LRU is a generic least-recently-used cache with a fixed capacity.
// It is not safe for concurrent use; callers must synchronise access.
type LRU[K comparable, V any] struct {
	cap   int
	items map[K]*list.Element
	order *list.List
}

type lruEntry[K comparable, V any] struct {
	key K
	val V
}

// NewLRU creates an LRU cache that holds at most capacity items.
// It panics if capacity is less than 1.
func NewLRU[K comparable, V any](capacity int) *LRU[K, V] {
	if capacity < 1 {
		panic("lru: capacity must be >= 1")
	}
	return &LRU[K, V]{
		cap:   capacity,
		items: make(map[K]*list.Element, capacity),
		order: list.New(),
	}
}

// Get retrieves the value for key and marks it as recently used.
// It returns the zero value and false if the key is absent.
func (c *LRU[K, V]) Get(key K) (V, bool) {
	el, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	c.order.MoveToFront(el)
	return el.Value.(*lruEntry[K, V]).val, true
}

// Put adds or updates a key-value pair. If the cache is at capacity the
// least-recently-used item is evicted.
func (c *LRU[K, V]) Put(key K, value V) {
	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		el.Value.(*lruEntry[K, V]).val = value
		return
	}
	if c.order.Len() >= c.cap {
		c.evict()
	}
	entry := &lruEntry[K, V]{key: key, val: value}
	el := c.order.PushFront(entry)
	c.items[key] = el
}

// Delete removes a key from the cache. It is a no-op if key is absent.
func (c *LRU[K, V]) Delete(key K) {
	el, ok := c.items[key]
	if !ok {
		return
	}
	c.order.Remove(el)
	delete(c.items, key)
}

// Len returns the current number of items in the cache.
func (c *LRU[K, V]) Len() int {
	return c.order.Len()
}

// Has reports whether key exists in the cache without updating recency.
func (c *LRU[K, V]) Has(key K) bool {
	_, ok := c.items[key]
	return ok
}

// Clear removes all items from the cache.
func (c *LRU[K, V]) Clear() {
	c.items = make(map[K]*list.Element, c.cap)
	c.order.Init()
}

// Keys returns all keys in order from most to least recently used.
func (c *LRU[K, V]) Keys() []K {
	keys := make([]K, 0, c.order.Len())
	for el := c.order.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*lruEntry[K, V]).key)
	}
	return keys
}

func (c *LRU[K, V]) evict() {
	el := c.order.Back()
	if el == nil {
		return
	}
	c.order.Remove(el)
	delete(c.items, el.Value.(*lruEntry[K, V]).key)
}
