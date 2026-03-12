package types

import (
	"sync"
	"testing"
	"time"
)

func TestTTLCacheSetGet(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	c.Set("a", 1)
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatalf("Get(a) = %d, %v; want 1, true", v, ok)
	}
}

func TestTTLCacheGetMissing(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("Get(missing) should return false")
	}
}

func TestTTLCacheExpiration(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](50 * time.Millisecond)
	c.Set("a", 1)
	time.Sleep(100 * time.Millisecond)
	_, ok := c.Get("a")
	if ok {
		t.Fatal("expired entry should not be returned")
	}
}

func TestTTLCacheSetWithTTL(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	c.SetWithTTL("short", 1, 50*time.Millisecond)
	c.SetWithTTL("long", 2, time.Hour)

	time.Sleep(100 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Fatal("short TTL entry should have expired")
	}
	v, ok := c.Get("long")
	if !ok || v != 2 {
		t.Fatalf("long TTL entry should still exist: %d, %v", v, ok)
	}
}

func TestTTLCacheDelete(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	c.Set("a", 1)
	c.Delete("a")
	_, ok := c.Get("a")
	if ok {
		t.Fatal("deleted entry should not be returned")
	}
}

func TestTTLCacheLen(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	if c.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", c.Len())
	}
	c.Set("a", 1)
	c.Set("b", 2)
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
}

func TestTTLCachePurge(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](50 * time.Millisecond)
	c.Set("a", 1)
	c.Set("b", 2)
	c.SetWithTTL("c", 3, time.Hour)

	time.Sleep(100 * time.Millisecond)

	removed := c.Purge()
	if removed != 2 {
		t.Fatalf("Purge() removed %d, want 2", removed)
	}
	if c.Len() != 1 {
		t.Fatalf("Len() after purge = %d, want 1", c.Len())
	}
}

func TestTTLCacheClear(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	c.Set("a", 1)
	c.Set("b", 2)
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("Len() after clear = %d, want 0", c.Len())
	}
}

func TestTTLCacheKeys(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](50 * time.Millisecond)
	c.Set("expired", 1)
	c.SetWithTTL("alive", 2, time.Hour)

	time.Sleep(100 * time.Millisecond)

	keys := c.Keys()
	if len(keys) != 1 || keys[0] != "alive" {
		t.Fatalf("Keys() = %v, want [alive]", keys)
	}
}

func TestTTLCacheOverwrite(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[string, int](time.Hour)
	c.Set("a", 1)
	c.Set("a", 2)
	v, ok := c.Get("a")
	if !ok || v != 2 {
		t.Fatalf("Get(a) after overwrite = %d, %v; want 2, true", v, ok)
	}
}

func TestTTLCacheConcurrency(t *testing.T) {
	t.Parallel()
	c := NewTTLCache[int, int](time.Hour)
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n)
			c.Get(n)
			c.Delete(n)
		}(i)
	}
	wg.Wait()
}

func BenchmarkTTLCacheSet(b *testing.B) {
	c := NewTTLCache[int, int](time.Hour)
	for b.Loop() {
		c.Set(1, 1)
	}
}

func BenchmarkTTLCacheGet(b *testing.B) {
	c := NewTTLCache[int, int](time.Hour)
	c.Set(1, 1)
	b.ResetTimer()
	for b.Loop() {
		c.Get(1)
	}
}
