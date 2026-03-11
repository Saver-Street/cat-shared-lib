package types

import (
	"sort"
	"sync"
	"testing"
)

func TestSyncMapSetAndGet(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("b", 2)

	v, ok := sm.Get("a")
	if !ok || v != 1 {
		t.Errorf("Get(a) = %d, %v; want 1, true", v, ok)
	}
	v, ok = sm.Get("b")
	if !ok || v != 2 {
		t.Errorf("Get(b) = %d, %v; want 2, true", v, ok)
	}
	_, ok = sm.Get("c")
	if ok {
		t.Error("Get(c) should return false")
	}
}

func TestSyncMapOverwrite(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("a", 42)

	v, ok := sm.Get("a")
	if !ok || v != 42 {
		t.Errorf("Get(a) = %d, %v; want 42, true", v, ok)
	}
}

func TestSyncMapDelete(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Delete("a")

	_, ok := sm.Get("a")
	if ok {
		t.Error("Get(a) should return false after delete")
	}

	// Deleting non-existent key should not panic.
	sm.Delete("nonexistent")
}

func TestSyncMapHas(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("x", 10)

	if !sm.Has("x") {
		t.Error("Has(x) = false; want true")
	}
	if sm.Has("y") {
		t.Error("Has(y) = true; want false")
	}
}

func TestSyncMapLen(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	if sm.Len() != 0 {
		t.Errorf("Len() = %d; want 0", sm.Len())
	}
	sm.Set("a", 1)
	sm.Set("b", 2)
	if sm.Len() != 2 {
		t.Errorf("Len() = %d; want 2", sm.Len())
	}
	sm.Delete("a")
	if sm.Len() != 1 {
		t.Errorf("Len() = %d; want 1", sm.Len())
	}
}

func TestSyncMapKeys(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("b", 2)
	sm.Set("a", 1)
	sm.Set("c", 3)

	keys := sm.Keys()
	sort.Strings(keys)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("Keys() = %v; want [a b c]", keys)
	}
}

func TestSyncMapValues(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("b", 2)

	vals := sm.Values()
	sort.Ints(vals)
	if len(vals) != 2 || vals[0] != 1 || vals[1] != 2 {
		t.Errorf("Values() = %v; want [1 2]", vals)
	}
}

func TestSyncMapRange(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("b", 2)
	sm.Set("c", 3)

	var count int
	sm.Range(func(k string, v int) bool {
		count++
		return true
	})
	if count != 3 {
		t.Errorf("Range visited %d; want 3", count)
	}
}

func TestSyncMapRangeEarlyStop(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("b", 2)
	sm.Set("c", 3)

	var count int
	sm.Range(func(k string, v int) bool {
		count++
		return false
	})
	if count != 1 {
		t.Errorf("Range visited %d; want 1 (early stop)", count)
	}
}

func TestSyncMapGetOrSet(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()

	// First call stores.
	v, loaded := sm.GetOrSet("a", 1)
	if loaded || v != 1 {
		t.Errorf("GetOrSet(a,1) = %d, %v; want 1, false", v, loaded)
	}

	// Second call loads existing.
	v, loaded = sm.GetOrSet("a", 99)
	if !loaded || v != 1 {
		t.Errorf("GetOrSet(a,99) = %d, %v; want 1, true", v, loaded)
	}
}

func TestSyncMapClear(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[string, int]()
	sm.Set("a", 1)
	sm.Set("b", 2)
	sm.Clear()

	if sm.Len() != 0 {
		t.Errorf("Len() = %d after Clear; want 0", sm.Len())
	}
	if sm.Has("a") {
		t.Error("Has(a) = true after Clear; want false")
	}
}

func TestSyncMapConcurrent(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[int, int]()
	const goroutines = 100
	const ops = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(id int) {
			defer wg.Done()
			for j := range ops {
				key := id*ops + j
				sm.Set(key, key)
				sm.Get(key)
				sm.Has(key)
				sm.Len()
			}
		}(i)
	}
	wg.Wait()

	if sm.Len() != goroutines*ops {
		t.Errorf("Len() = %d; want %d", sm.Len(), goroutines*ops)
	}
}

func TestSyncMapIntKeys(t *testing.T) {
	t.Parallel()
	sm := NewSyncMap[int, string]()
	sm.Set(1, "one")
	sm.Set(2, "two")

	v, ok := sm.Get(1)
	if !ok || v != "one" {
		t.Errorf("Get(1) = %q, %v; want one, true", v, ok)
	}
}

func BenchmarkSyncMapSet(b *testing.B) {
	sm := NewSyncMap[int, int]()
	for i := range b.N {
		sm.Set(i, i)
	}
}

func BenchmarkSyncMapGet(b *testing.B) {
	sm := NewSyncMap[int, int]()
	for i := range 1000 {
		sm.Set(i, i)
	}
	b.ResetTimer()
	for i := range b.N {
		sm.Get(i % 1000)
	}
}

func BenchmarkSyncMapConcurrent(b *testing.B) {
	sm := NewSyncMap[int, int]()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			sm.Set(i, i)
			sm.Get(i)
			i++
		}
	})
}

func FuzzSyncMapSetGet(f *testing.F) {
	f.Add("key", 42)
	f.Add("", 0)
	f.Add("hello world", -1)
	f.Fuzz(func(t *testing.T, key string, value int) {
		sm := NewSyncMap[string, int]()
		sm.Set(key, value)
		got, ok := sm.Get(key)
		if !ok {
			t.Errorf("Get(%q) not found after Set", key)
		}
		if got != value {
			t.Errorf("Get(%q) = %d; want %d", key, got, value)
		}
	})
}
