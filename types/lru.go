package types

import "container/list"

// LRUCache is a bounded Least Recently Used cache.
// When the cache reaches capacity, the least recently accessed entry
// is evicted. The zero value is not usable; call NewLRUCache.
type LRUCache[K comparable, V any] struct {
	cap   int
	ll    *list.List
	items map[K]*list.Element
}

type lruEntry[K comparable, V any] struct {
	key K
	val V
}

// NewLRUCache returns an LRUCache with the given capacity.
// Panics if capacity < 1.
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	if capacity < 1 {
		panic("lru: capacity must be >= 1")
	}
	return &LRUCache[K, V]{
		cap:   capacity,
		ll:    list.New(),
		items: make(map[K]*list.Element, capacity),
	}
}

// Get returns the value for key and moves it to the front.
// The bool is false if the key is not present.
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		return el.Value.(*lruEntry[K, V]).val, true
	}
	var zero V
	return zero, false
}

// Put inserts or updates the entry for key and moves it to the front.
// If the cache is full, the least recently used entry is evicted.
func (c *LRUCache[K, V]) Put(key K, value V) {
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		el.Value.(*lruEntry[K, V]).val = value
		return
	}
	if c.ll.Len() >= c.cap {
		c.evict()
	}
	el := c.ll.PushFront(&lruEntry[K, V]{key: key, val: value})
	c.items[key] = el
}

// Delete removes the entry for key.
func (c *LRUCache[K, V]) Delete(key K) {
	if el, ok := c.items[key]; ok {
		c.removeElement(el)
	}
}

// Contains reports whether key is present without changing recency.
func (c *LRUCache[K, V]) Contains(key K) bool {
	_, ok := c.items[key]
	return ok
}

// Len returns the number of entries in the cache.
func (c *LRUCache[K, V]) Len() int {
	return c.ll.Len()
}

// Cap returns the maximum number of entries the cache can hold.
func (c *LRUCache[K, V]) Cap() int {
	return c.cap
}

// Clear removes all entries from the cache.
func (c *LRUCache[K, V]) Clear() {
	c.ll.Init()
	c.items = make(map[K]*list.Element, c.cap)
}

// Keys returns all keys in order from most recently used to least.
func (c *LRUCache[K, V]) Keys() []K {
	keys := make([]K, 0, c.ll.Len())
	for el := c.ll.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*lruEntry[K, V]).key)
	}
	return keys
}

func (c *LRUCache[K, V]) evict() {
	el := c.ll.Back()
	if el != nil {
		c.removeElement(el)
	}
}

func (c *LRUCache[K, V]) removeElement(el *list.Element) {
	c.ll.Remove(el)
	delete(c.items, el.Value.(*lruEntry[K, V]).key)
}
