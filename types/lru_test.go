package types

import "testing"

func TestLRUGetPut(t *testing.T) {
	c := NewLRU[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatalf("Get(a) = %d, %v; want 1, true", v, ok)
	}
	_, ok = c.Get("z")
	if ok {
		t.Fatal("Get(z) should return false")
	}
}

func TestLRUEviction(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3) // evicts "a"

	if c.Has("a") {
		t.Fatal("a should be evicted")
	}
	v, ok := c.Get("b")
	if !ok || v != 2 {
		t.Fatal("b should still exist")
	}
	v, ok = c.Get("c")
	if !ok || v != 3 {
		t.Fatal("c should exist")
	}
}

func TestLRUAccessUpdatesRecency(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Get("a")    // make "a" recent
	c.Put("c", 3) // should evict "b" not "a"

	if c.Has("b") {
		t.Fatal("b should be evicted (least recent)")
	}
	if !c.Has("a") {
		t.Fatal("a should still exist (recently accessed)")
	}
}

func TestLRUUpdateExisting(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("a", 10) // update "a"

	v, _ := c.Get("a")
	if v != 10 {
		t.Fatalf("Get(a) = %d, want 10", v)
	}
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
}

func TestLRUDelete(t *testing.T) {
	c := NewLRU[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Delete("a")

	if c.Has("a") {
		t.Fatal("a should be deleted")
	}
	if c.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", c.Len())
	}
}

func TestLRUDeleteMissing(t *testing.T) {
	c := NewLRU[string, int](2)
	c.Put("a", 1)
	c.Delete("missing") // should not panic
	if c.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", c.Len())
	}
}

func TestLRULen(t *testing.T) {
	c := NewLRU[int, int](5)
	if c.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", c.Len())
	}
	c.Put(1, 1)
	c.Put(2, 2)
	if c.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", c.Len())
	}
}

func TestLRUHas(t *testing.T) {
	c := NewLRU[string, string](2)
	c.Put("key", "val")
	if !c.Has("key") {
		t.Fatal("Has(key) should be true")
	}
	if c.Has("nope") {
		t.Fatal("Has(nope) should be false")
	}
}

func TestLRUClear(t *testing.T) {
	c := NewLRU[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Clear()

	if c.Len() != 0 {
		t.Fatalf("Len() after Clear = %d, want 0", c.Len())
	}
	if c.Has("a") {
		t.Fatal("a should not exist after Clear")
	}
}

func TestLRUKeys(t *testing.T) {
	c := NewLRU[string, int](3)
	c.Put("a", 1)
	c.Put("b", 2)
	c.Put("c", 3)
	c.Get("a") // make "a" most recent

	keys := c.Keys()
	// Most recent first: a, c, b
	if len(keys) != 3 {
		t.Fatalf("Keys() len = %d, want 3", len(keys))
	}
	if keys[0] != "a" {
		t.Fatalf("Keys()[0] = %q, want a (most recent)", keys[0])
	}
}

func TestLRUPanicOnZeroCapacity(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for capacity 0")
		}
	}()
	NewLRU[int, int](0)
}

func BenchmarkLRUPut(b *testing.B) {
	c := NewLRU[int, int](100)
	for b.Loop() {
		for i := range 200 {
			c.Put(i, i)
		}
	}
}

func BenchmarkLRUGet(b *testing.B) {
	c := NewLRU[int, int](100)
	for i := range 100 {
		c.Put(i, i)
	}
	b.ResetTimer()
	for b.Loop() {
		c.Get(50)
	}
}

func FuzzLRUPutGet(f *testing.F) {
	f.Add("key", 42)
	f.Add("", 0)

	f.Fuzz(func(t *testing.T, k string, v int) {
		c := NewLRU[string, int](10)
		c.Put(k, v)
		got, ok := c.Get(k)
		if !ok {
			t.Fatal("key not found after Put")
		}
		if got != v {
			t.Fatalf("Get(%q) = %d, want %d", k, got, v)
		}
	})
}
