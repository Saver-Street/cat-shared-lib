package types

import (
	"testing"
)

func TestMultiMapNewEmpty(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	if mm.Len() != 0 {
		t.Errorf("Len = %d; want 0", mm.Len())
	}
}

func TestMultiMapPutGet(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1, 2)
	mm.Put("a", 3)
	vals, ok := mm.Get("a")
	if !ok || len(vals) != 3 {
		t.Fatalf("Get(a) = %v, %v; want [1 2 3], true", vals, ok)
	}
	if vals[0] != 1 || vals[1] != 2 || vals[2] != 3 {
		t.Errorf("values = %v; want [1 2 3]", vals)
	}
}

func TestMultiMapGetMissing(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	vals, ok := mm.Get("missing")
	if ok || vals != nil {
		t.Errorf("Get(missing) = %v, %v; want nil, false", vals, ok)
	}
}

func TestMultiMapZeroValue(t *testing.T) {
	t.Parallel()
	var mm MultiMap[string, int]
	// zero value should work without init
	mm.Put("x", 1)
	vals, ok := mm.Get("x")
	if !ok || len(vals) != 1 {
		t.Fatalf("zero value Put/Get = %v, %v; want [1], true", vals, ok)
	}
}

func TestMultiMapZeroValueGet(t *testing.T) {
	t.Parallel()
	var mm MultiMap[string, int]
	vals, ok := mm.Get("x")
	if ok || vals != nil {
		t.Errorf("zero Get = %v, %v; want nil, false", vals, ok)
	}
}

func TestMultiMapZeroValueContains(t *testing.T) {
	t.Parallel()
	var mm MultiMap[string, int]
	if mm.Contains("x") {
		t.Error("zero Contains = true; want false")
	}
}

func TestMultiMapZeroValueDelete(t *testing.T) {
	t.Parallel()
	var mm MultiMap[string, int]
	mm.Delete("x") // should not panic
}

func TestMultiMapDelete(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1)
	mm.Delete("a")
	if mm.Contains("a") {
		t.Error("Contains(a) after Delete = true")
	}
	if mm.Len() != 0 {
		t.Errorf("Len = %d after Delete; want 0", mm.Len())
	}
}

func TestMultiMapContains(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1)
	if !mm.Contains("a") {
		t.Error("Contains(a) = false; want true")
	}
	if mm.Contains("b") {
		t.Error("Contains(b) = true; want false")
	}
}

func TestMultiMapLen(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1, 2)
	mm.Put("b", 3)
	if mm.Len() != 2 {
		t.Errorf("Len = %d; want 2", mm.Len())
	}
}

func TestMultiMapValueCount(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1, 2)
	mm.Put("b", 3)
	if mm.ValueCount() != 3 {
		t.Errorf("ValueCount = %d; want 3", mm.ValueCount())
	}
}

func TestMultiMapKeys(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1)
	mm.Put("b", 2)
	keys := mm.Keys()
	if len(keys) != 2 {
		t.Fatalf("Keys len = %d; want 2", len(keys))
	}
	seen := map[string]bool{}
	for _, k := range keys {
		seen[k] = true
	}
	if !seen["a"] || !seen["b"] {
		t.Errorf("Keys = %v; want {a, b}", keys)
	}
}

func TestMultiMapKeysNil(t *testing.T) {
	t.Parallel()
	var mm MultiMap[string, int]
	keys := mm.Keys()
	if keys != nil {
		t.Errorf("nil Keys = %v; want nil", keys)
	}
}

func TestMultiMapEach(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1, 2)
	mm.Put("b", 3)
	sum := 0
	mm.Each(func(_ string, v int) bool {
		sum += v
		return true
	})
	if sum != 6 {
		t.Errorf("Each sum = %d; want 6", sum)
	}
}

func TestMultiMapEachEarlyStop(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1, 2, 3, 4, 5)
	count := 0
	mm.Each(func(_ string, _ int) bool {
		count++
		return count < 3
	})
	if count != 3 {
		t.Errorf("Each stopped at count = %d; want 3", count)
	}
}

func TestMultiMapClear(t *testing.T) {
	t.Parallel()
	mm := NewMultiMap[string, int]()
	mm.Put("a", 1)
	mm.Put("b", 2)
	mm.Clear()
	if mm.Len() != 0 {
		t.Errorf("Len after Clear = %d; want 0", mm.Len())
	}
}

func BenchmarkMultiMapPut(b *testing.B) {
	mm := NewMultiMap[int, int]()
	for i := range b.N {
		mm.Put(i%100, i)
	}
}

func BenchmarkMultiMapGet(b *testing.B) {
	mm := NewMultiMap[int, int]()
	for i := range 1000 {
		mm.Put(i, i)
	}
	b.ResetTimer()
	for i := range b.N {
		mm.Get(i % 1000)
	}
}

func FuzzMultiMapPutGet(f *testing.F) {
	f.Add("key", 42)
	f.Add("", 0)
	f.Add("x", -1)
	f.Fuzz(func(t *testing.T, key string, val int) {
		mm := NewMultiMap[string, int]()
		mm.Put(key, val)
		vals, ok := mm.Get(key)
		if !ok || len(vals) != 1 || vals[0] != val {
			t.Errorf("Put/Get(%q, %d) = %v, %v", key, val, vals, ok)
		}
	})
}
