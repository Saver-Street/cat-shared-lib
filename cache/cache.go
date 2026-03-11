// Package cache provides a thread-safe, generic in-memory LRU cache with
// per-entry TTL support and automatic expired-entry cleanup.
package cache

import (
	"container/list"
	"sync"
	"time"
)

// Config configures the LRU cache.
type Config struct {
	// MaxEntries is the maximum number of entries. Default: 1000.
	MaxEntries int
	// DefaultTTL is the default time-to-live for entries. Default: 5 minutes.
	// Set to 0 for no expiration.
	DefaultTTL time.Duration
	// CleanupInterval is how often expired entries are swept. Default: 1 minute.
	// Set to 0 to disable background cleanup.
	CleanupInterval time.Duration
}

func (c *Config) defaults() {
	if c.MaxEntries <= 0 {
		c.MaxEntries = 1000
	}
	if c.DefaultTTL < 0 {
		c.DefaultTTL = 5 * time.Minute
	}
	if c.DefaultTTL == 0 && c.CleanupInterval == 0 {
		// No TTL and no cleanup requested — leave as is.
	} else if c.CleanupInterval < 0 {
		c.CleanupInterval = time.Minute
	}
}

type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

// Cache is a generic LRU cache with optional TTL.
type Cache[K comparable, V any] struct {
	config  Config
	mu      sync.RWMutex
	items   map[K]*list.Element
	order   *list.List
	stopCh  chan struct{}
	stopped bool
	now     func() time.Time
}

// New creates a new LRU cache. Call Stop when the cache is no longer needed
// to halt the background cleanup goroutine.
func New[K comparable, V any](cfg Config) *Cache[K, V] {
	cfg.defaults()
	c := &Cache[K, V]{
		config: cfg,
		items:  make(map[K]*list.Element, cfg.MaxEntries),
		order:  list.New(),
		stopCh: make(chan struct{}),
		now:    time.Now,
	}
	if cfg.CleanupInterval > 0 {
		go c.cleanupLoop()
	}
	return c
}

// Set adds or updates a cache entry with the default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.config.DefaultTTL)
}

// SetWithTTL adds or updates a cache entry with a specific TTL.
// A TTL of 0 means the entry never expires.
func (c *Cache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := c.now()
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = now.Add(ttl)
	}

	if el, ok := c.items[key]; ok {
		c.order.MoveToFront(el)
		e := el.Value.(*entry[K, V])
		e.value = value
		e.expiresAt = expiresAt
		return
	}

	e := &entry[K, V]{key: key, value: value, expiresAt: expiresAt}
	el := c.order.PushFront(e)
	c.items[key] = el

	if c.order.Len() > c.config.MaxEntries {
		c.evictOldest()
	}
}

// Get retrieves a value from the cache. The second return value indicates
// whether the key was found and not expired.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}

	e := el.Value.(*entry[K, V])
	if !e.expiresAt.IsZero() && c.now().After(e.expiresAt) {
		c.removeElement(el)
		var zero V
		return zero, false
	}

	c.order.MoveToFront(el)
	return e.value, true
}

// Contains reports whether key exists in the cache and has not expired.
// Unlike Get, it does not update the LRU order.
func (c *Cache[K, V]) Contains(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	el, ok := c.items[key]
	if !ok {
		return false
	}

	e := el.Value.(*entry[K, V])
	if !e.expiresAt.IsZero() && c.now().After(e.expiresAt) {
		c.removeElement(el)
		return false
	}

	return true
}

// Delete removes an entry from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if el, ok := c.items[key]; ok {
		c.removeElement(el)
	}
}

// Len returns the number of entries in the cache.
func (c *Cache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.order.Len()
}

// Clear removes all entries from the cache.
func (c *Cache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*list.Element, c.config.MaxEntries)
	c.order.Init()
}

// Keys returns a snapshot of all keys currently in the cache.
// The order is from most recently used to least recently used.
func (c *Cache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()
	keys := make([]K, 0, c.order.Len())
	for el := c.order.Front(); el != nil; el = el.Next() {
		keys = append(keys, el.Value.(*entry[K, V]).key)
	}
	return keys
}

// Stop halts the background cleanup goroutine. Safe to call multiple times.
func (c *Cache[K, V]) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.stopped {
		c.stopped = true
		close(c.stopCh)
	}
}

func (c *Cache[K, V]) evictOldest() {
	el := c.order.Back()
	if el != nil {
		c.removeElement(el)
	}
}

func (c *Cache[K, V]) removeElement(el *list.Element) {
	e := el.Value.(*entry[K, V])
	delete(c.items, e.key)
	c.order.Remove(el)
}

func (c *Cache[K, V]) cleanupLoop() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	defer ticker.Stop()
	for {
		select {
		case <-c.stopCh:
			return
		case <-ticker.C:
			c.removeExpired()
		}
	}
}

func (c *Cache[K, V]) removeExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	for el := c.order.Back(); el != nil; {
		prev := el.Prev()
		e := el.Value.(*entry[K, V])
		if !e.expiresAt.IsZero() && now.After(e.expiresAt) {
			c.removeElement(el)
		}
		el = prev
	}
}

// GetOrSet returns the cached value for key if present. If absent, it calls
// fill to compute the value, stores it with the default TTL, and returns it.
// The fill function is called while the cache lock is NOT held to avoid
// blocking other cache operations during slow computations.
func (c *Cache[K, V]) GetOrSet(key K, fill func() V) V {
if v, ok := c.Get(key); ok {
return v
}
v := fill()
c.Set(key, v)
return v
}
