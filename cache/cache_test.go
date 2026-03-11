package cache

import (
	"sync"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNew_Defaults(t *testing.T) {
	c := New[string, int](Config{})
	defer c.Stop()

	testkit.AssertEqual(t, c.config.MaxEntries, 1000)
}

func TestSetAndGet(t *testing.T) {
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", "value1")
	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	testkit.AssertEqual(t, v, "value1")
}

func TestGet_NotFound(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10})
	defer c.Stop()

	_, ok := c.Get("missing")
	testkit.AssertFalse(t, ok)
}

func TestSet_UpdateExisting(t *testing.T) {
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", "v1")
	c.Set("key1", "v2")

	v, ok := c.Get("key1")
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, v, "v2")
	testkit.AssertEqual(t, c.Len(), 1)
}

func TestLRU_Eviction(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 3, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Set("d", 4) // should evict "a"

	_, ok := c.Get("a")
	testkit.AssertFalse(t, ok)
	testkit.AssertEqual(t, c.Len(), 3)
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

	_, ok := c.Get("a")
	testkit.AssertTrue(t, ok)
	_, ok = c.Get("b")
	testkit.AssertFalse(t, ok)
}

func TestTTL_Expiration(t *testing.T) {
	now := time.Now()
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Second})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.Set("key1", "value1")

	// Before expiry.
	_, ok := c.Get("key1")
	testkit.AssertTrue(t, ok)

	// After expiry.
	now = now.Add(2 * time.Second)
	_, ok = c.Get("key1")
	testkit.AssertFalse(t, ok)
	testkit.AssertEqual(t, c.Len(), 0)
}

func TestSetWithTTL_NoExpiration(t *testing.T) {
	now := time.Now()
	c := New[string, string](Config{MaxEntries: 10, DefaultTTL: time.Second, CleanupInterval: 0})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.SetWithTTL("forever", "value", 0) // No TTL.

	now = now.Add(time.Hour)
	v, ok := c.Get("forever")
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, v, "value")
}

func TestSetWithTTL_CustomTTL(t *testing.T) {
	now := time.Now()
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Hour})
	defer c.Stop()
	c.now = func() time.Time { return now }

	c.SetWithTTL("fast", 42, 100*time.Millisecond)

	now = now.Add(200 * time.Millisecond)
	_, ok := c.Get("fast")
	testkit.AssertFalse(t, ok)
}

func TestDelete(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
	defer c.Stop()

	c.Set("key1", 1)
	c.Delete("key1")

	_, ok := c.Get("key1")
	testkit.AssertFalse(t, ok)
	testkit.AssertEqual(t, c.Len(), 0)
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

	testkit.AssertEqual(t, c.Len(), 0)
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

	testkit.AssertEqual(t, c.Len(), 1)
	_, ok := c.Get("keep")
	testkit.AssertTrue(t, ok)
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

	testkit.AssertEqual(t, c.Len(), 0)
	c.Set("a", 1)
	c.Set("b", 2)
	testkit.AssertEqual(t, c.Len(), 2)
}

func TestNew_NoCleanup_NoDefaultTTL(t *testing.T) {
	c := New[string, int](Config{MaxEntries: 5, DefaultTTL: 0, CleanupInterval: 0})
	defer c.Stop()

	c.Set("a", 1)
	v, ok := c.Get("a")
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, v, 1)
}

func TestNew_NegativeCleanupInterval_Defaults(t *testing.T) {
	cfg := Config{DefaultTTL: time.Second, CleanupInterval: -1}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.CleanupInterval, time.Minute)
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

	testkit.AssertEqual(t, c.Len(), 0)
}

func TestNew_NegativeDefaultTTL_Defaults(t *testing.T) {
	cfg := Config{DefaultTTL: -1}
	cfg.defaults()
	testkit.AssertEqual(t, cfg.DefaultTTL, 5*time.Minute)
}

func BenchmarkSet(b *testing.B) {
	c := New[int, int](Config{MaxEntries: 10000, DefaultTTL: time.Minute, CleanupInterval: 0})
	defer c.Stop()
	b.ResetTimer()
	var i int
	for b.Loop() {
		i++
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
	var i int
	for b.Loop() {
		c.Get(i % 10000)
		i++
	}
}

func TestKeys_Empty(t *testing.T) {
c := New[string, int](Config{MaxEntries: 5, DefaultTTL: time.Minute})
defer c.Stop()
testkit.AssertLen(t, c.Keys(), 0)
}

func TestKeys_Order(t *testing.T) {
c := New[string, int](Config{MaxEntries: 5, DefaultTTL: time.Minute})
defer c.Stop()
c.Set("a", 1)
c.Set("b", 2)
c.Set("c", 3)

// Most recently used first: c, b, a
keys := c.Keys()
testkit.AssertLen(t, keys, 3)
testkit.AssertEqual(t, keys[0], "c")
testkit.AssertEqual(t, keys[1], "b")
testkit.AssertEqual(t, keys[2], "a")
}

func TestKeys_AfterAccess(t *testing.T) {
c := New[string, int](Config{MaxEntries: 5, DefaultTTL: time.Minute})
defer c.Stop()
c.Set("a", 1)
c.Set("b", 2)
c.Set("c", 3)

// Access "a" to move it to front
c.Get("a")

keys := c.Keys()
testkit.AssertEqual(t, keys[0], "a")
}

func TestGetOrSet_Miss(t *testing.T) {
c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
defer c.Stop()

calls := 0
v := c.GetOrSet("key", func() int {
calls++
return 42
})
testkit.AssertEqual(t, v, 42)
testkit.AssertEqual(t, calls, 1)

// Should be cached now
got, ok := c.Get("key")
testkit.AssertTrue(t, ok)
testkit.AssertEqual(t, got, 42)
}

func TestGetOrSet_Hit(t *testing.T) {
c := New[string, int](Config{MaxEntries: 10, DefaultTTL: time.Minute})
defer c.Stop()
c.Set("key", 100)

calls := 0
v := c.GetOrSet("key", func() int {
calls++
return 999
})
testkit.AssertEqual(t, v, 100) // should return cached value
testkit.AssertEqual(t, calls, 0) // fill should not be called
}

func TestContains_Hit(t *testing.T) {
c := New[string, int](Config{DefaultTTL: time.Minute})
defer c.Stop()

c.Set("k", 42)
testkit.AssertTrue(t, c.Contains("k"))
}

func TestContains_Miss(t *testing.T) {
c := New[string, int](Config{DefaultTTL: time.Minute})
defer c.Stop()

testkit.AssertFalse(t, c.Contains("nope"))
}

func TestContains_Expired(t *testing.T) {
now := time.Now()
c := New[string, int](Config{DefaultTTL: time.Minute})
defer c.Stop()
c.now = func() time.Time { return now }

c.Set("k", 1)
c.now = func() time.Time { return now.Add(2 * time.Minute) }
testkit.AssertFalse(t, c.Contains("k"))
}

func TestContains_DoesNotPromoteLRU(t *testing.T) {
c := New[string, int](Config{DefaultTTL: time.Minute, MaxEntries: 2})
defer c.Stop()

c.Set("a", 1)
c.Set("b", 2)
// Contains should NOT promote "a"
c.Contains("a")
// Adding a third should evict "a" (oldest) since Contains didn't promote it
c.Set("c", 3)
testkit.AssertFalse(t, c.Contains("a"))
testkit.AssertTrue(t, c.Contains("b"))
testkit.AssertTrue(t, c.Contains("c"))
}
