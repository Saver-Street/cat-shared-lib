package types

import "testing"

func TestLRUCacheNew(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](10)
	if c.Len() != 0 {
		t.Errorf("Len = %d; want 0", c.Len())
	}
	if c.Cap() != 10 {
		t.Errorf("Cap = %d; want 10", c.Cap())
	}
}

func TestLRUCacheNewPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for capacity 0")
		}
	}()
	NewLRUCache[string, int](0)
}

func TestLRUCachePutGet(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Errorf("Get(a) = %d, %v; want 1, true", v, ok)
	}
}

func TestLRUCacheGetMissing(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	v, ok := c.Get("missing")
	if ok || v != 0 {
		t.Errorf("Get(missing) = %d, %v; want 0, false", v, ok)
	}
}

func TestLRUCacheEviction(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"
	if c.Contains("a") {
		t.Error("Contains(a) after eviction = true")
	}
	if !c.Contains("b") || !c.Contains("c") {
		t.Error("b or c missing after eviction")
	}
}

func TestLRUCacheGetPromotes(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")    // promote a
	c.Put("c", 3) // should evict b, not a
	if !c.Contains("a") {
		t.Error("Contains(a) = false; Get should have promoted it")
	}
	if c.Contains("b") {
		t.Error("Contains(b) = true; should have been evicted")
	}
}

func TestLRUCachePutUpdate(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](2)
	c.Put("a", 1)
	c.Put("a", 2) // update
	v, ok := c.Get("a")
	if !ok || v != 2 {
		t.Errorf("Get(a) = %d, %v; want 2, true", v, ok)
	}
	if c.Len() != 1 {
		t.Errorf("Len = %d; want 1", c.Len())
	}
}

func TestLRUCacheDelete(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	c.Delete("a")
	if c.Contains("a") {
		t.Error("Contains(a) after Delete = true")
	}
	if c.Len() != 0 {
		t.Errorf("Len after Delete = %d; want 0", c.Len())
	}
}

func TestLRUCacheDeleteMissing(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Delete("missing") // should not panic
}

func TestLRUCacheContains(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	if !c.Contains("a") {
		t.Error("Contains(a) = false")
	}
	if c.Contains("b") {
		t.Error("Contains(b) = true")
	}
}

func TestLRUCacheClear(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Clear()
	if c.Len() != 0 {
		t.Errorf("Len after Clear = %d; want 0", c.Len())
	}
}

func TestLRUCacheKeys(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	keys := c.Keys()
	// Most recent first: c, b, a
	if len(keys) != 3 || keys[0] != "c" || keys[1] != "b" || keys[2] != "a" {
		t.Errorf("Keys = %v; want [c b a]", keys)
	}
}

func TestLRUCacheKeysAfterGet(t *testing.T) {
	t.Parallel()
	c := NewLRUCache[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Get("a") // promote a
	keys := c.Keys()
	if len(keys) != 3 || keys[0] != "a" {
		t.Errorf("Keys after Get = %v; want a first", keys)
	}
}

func BenchmarkLRUCachePut(b *testing.B) {
	c := NewLRUCache[int, int](1000)
	for i := range b.N {
		c.Put(i, i)
	}
}

func BenchmarkLRUCacheGet(b *testing.B) {
	c := NewLRUCache[int, int](1000)
	for i := range 1000 {
		c.Put(i, i)
	}
	b.ResetTimer()
	for i := range b.N {
		c.Get(i % 1000)
	}
}

func FuzzLRUCache(f *testing.F) {
	f.Add("key", 42)
	f.Add("", 0)
	f.Fuzz(func(t *testing.T, key string, val int) {
		c := NewLRUCache[string, int](5)
		c.Put(key, val)
		got, ok := c.Get(key)
		if !ok || got != val {
			t.Errorf("Put/Get(%q, %d) = %d, %v", key, val, got, ok)
		}
	})
}
