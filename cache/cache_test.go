package cache

import (
	"sync"
	"testing"
	"time"
)

func TestNew_Defaults(t *testing.T) {
	c := New[string, int](Config{})
	defer c.Stop()

	if c.config.MaxEntries != 1000 {
		t.Errorf("expected 1000 max entries, got %d", c.config.MaxEntries)
	}
}

func TestSetAndGet(t *testing.T) {
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if v != "value1" {
		t.Errorf("expected value1, got %q", v)
	}
}

func TestGet_NotFound(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10})
	defer c.Stop()

	_, ok := c.Get("missing")
	if ok {
		t.Error("expected false for missing key")
	}
}

func TestSet_UpdateExisting(t *testing.T) {
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", "v1")
	c.Set("key1", "v2")

	v, ok := c.Get("key1")
	if !ok || v != "v2" {
		t.Errorf("expected updated value v2, got %q", v)
	}
	if c.Len() != 1 {
		t.Errorf("expected 1 entry after update, got %d", c.Len())
	}
}

func TestLRU_Eviction(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 3, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4) // should evict "a"

	if _, ok := c.Get("a"); ok {
		t.Error("expected 'a' to be evicted")
	}
	if c.Len() != 3 {
		t.Errorf("expected 3 entries, got %d", c.Len())
	}
}

func TestLRU_AccessPromotes(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 3, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)

	// Access "a" to promote it.
	c.Get("a")

	c.Set("d", 4) // should evict "b" (least recently used)

	if _, ok := c.Get("a"); !ok {
		t.Error("expected 'a' to still exist after promotion")
	}
	if _, ok := c.Get("b"); ok {
		t.Error("expected 'b' to be evicted")
	}
}

func TestTTL_Expiration(t *testing.T) {
	now := time.Now()
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Second})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.Set("key1", "value1")

	// Before expiry.
	if _, ok := c.Get("key1"); !ok {
		t.Error("expected key1 before TTL")
	}

	// After expiry.
	now = now.Add(2 * time.Second)
	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
	if c.Len() != 0 {
		t.Errorf("expected 0 entries after expiry, got %d", c.Len())
	}
}

func TestSetWithTTL_NoExpiration(t *testing.T) {
	now := time.Now()
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Second, CleanupInterval: 0})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.SetWithTTL("forever", "value", 0) // No TTL.

	now = now.Add(time.Hour)
	v, ok := c.Get("forever")
	if !ok {
		t.Error("entry with no TTL should never expire")
	}
	if v != "value" {
		t.Errorf("expected value, got %q", v)
	}
}

func TestSetWithTTL_CustomTTL(t *testing.T) {
	now := time.Now()
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Hour})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.SetWithTTL("fast", 42, 100*time.Millisecond)

	now = now.Add(200 * time.Millisecond)
	_, ok := c.Get("fast")
	if ok {
		t.Error("expected entry to be expired with custom short TTL")
	}
}

func TestDelete(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", 1)
	c.Delete("key1")

	if _, ok := c.Get("key1"); ok {
		t.Error("expected key1 to be deleted")
	}
	if c.Len() != 0 {
		t.Errorf("expected 0 entries, got %d", c.Len())
	}
}

func TestDelete_NonExistent(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10})
	defer c.Stop()
	c.Delete("nope") // should not panic
}

func TestClear(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()

	if c.Len() != 0 {
		t.Errorf("expected 0 after clear, got %d", c.Len())
	}
}

func TestStop_Idempotent(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10})
	c.Stop()
	c.Stop() // should not panic
}

func TestRemoveExpired(t *testing.T) {
	now := time.Now()
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Second, CleanupInterval: 0})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.Set("expire1", 1)
	c.Set("expire2", 2)
	c.SetWithTTL("keep", 3, 0) // No TTL.

	now = now.Add(2 * time.Second)
	c.removeExpired()

	if c.Len() != 1 {
		t.Errorf("expected 1 entry after cleanup, got %d", c.Len())
	}
	if _, ok := c.Get("keep"); !ok {
		t.Error("expected 'keep' to survive cleanup")
	}
}

func TestConcurrent_Access(t *testing.T) {
	c := New[int, int](Config{MaxEntries: 100, DefaultTTL: time.Minute})
	defer c.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n*2)
			c.Get(n)
			if n%3 == 0 {
				c.Delete(n)
			}
		}(i)
	}
	wg.Wait()
}

func TestLen(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	if c.Len() != 0 {
		t.Errorf("expected 0, got %d", c.Len())
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Errorf("expected 2, got %d", c.Len())
	}
}

func TestNew_NoCleanup_NoDefaultTTL(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 5, DefaultTTL: 0, CleanupInterval: 0})
	defer c.Stop()

	c.Set("a", 1)
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Error("expected entry to be retrievable")
	}
}

func TestNew_NegativeCleanupInterval_Defaults(t *testing.T) {
	cfg := Config{DefaultTTL: time.Second, CleanupInterval: -1}
	cfg.defaults()
	if cfg.CleanupInterval != time.Minute {
		t.Errorf("expected 1m default cleanup, got %v", cfg.CleanupInterval)
	}
}

func TestCleanupLoop_RunsAutomatically(t *testing.T) {
	now := time.Now()
	c := New[string, int](Config{
		MaxEntries:      10,
		DefaultTTL:      50 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
	})
	defer c.Stop()
	c.mu.Lock()
	c.now = func() time.Time { return now }
	c.mu.Unlock()

	c.Set("a", 1)

	// Advance time past TTL.
	c.mu.Lock()
	now = now.Add(100 * time.Millisecond)
	c.mu.Unlock()

	// Wait for cleanup loop to fire.
	time.Sleep(150 * time.Millisecond)

	if c.Len() != 0 {
		t.Errorf("expected 0 entries after cleanup loop, got %d", c.Len())
	}
}

func TestNew_NegativeDefaultTTL_Defaults(t *testing.T) {
	cfg := Config{DefaultTTL: -1}
	cfg.defaults()
	if cfg.DefaultTTL != 5*time.Minute {
		t.Errorf("expected 5m default TTL for negative value, got %v", cfg.DefaultTTL)
	}
}

func BenchmarkSet(b *testing.B) {
	c := New[int, int](Config{MaxEntries: 10000, DefaultTTL: time.Minute, CleanupInterval: 0})
	defer c.Stop()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Set(i, i)
	}
}

func BenchmarkGet(b *testing.B) {
	c := New[int, int](Config{MaxEntries: 10000, DefaultTTL: time.Minute, CleanupInterval: 0})
	defer c.Stop()
	for i := 0; i < 10000; i++ {
		c.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Get(i % 10000)
	}
}
