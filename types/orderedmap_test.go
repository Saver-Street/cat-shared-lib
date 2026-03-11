package types

import "testing"

func TestOrderedMapSetGet(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	v, ok := m.Get("b")
	if !ok || v != 2 {
		t.Fatalf("Get(b) = %d, %v; want 2, true", v, ok)
	}

	_, ok = m.Get("z")
	if ok {
		t.Fatal("Get(z) should return false")
	}
}

func TestOrderedMapUpdateRetainsOrder(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("a", 10) // update, should keep position

	keys := m.Keys()
	want := []string{"a", "b", "c"}
	if len(keys) != len(want) {
		t.Fatalf("keys = %v, want %v", keys, want)
	}
	for i := range want {
		if keys[i] != want[i] {
			t.Fatalf("keys[%d] = %q, want %q", i, keys[i], want[i])
		}
	}

	v, _ := m.Get("a")
	if v != 10 {
		t.Fatalf("Get(a) = %d, want 10", v)
	}
}

func TestOrderedMapDelete(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("x", 1)
	m.Set("y", 2)
	m.Set("z", 3)

	m.Delete("y")
	if m.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", m.Len())
	}
	if m.Has("y") {
		t.Fatal("Has(y) should be false after delete")
	}
	keys := m.Keys()
	if len(keys) != 2 || keys[0] != "x" || keys[1] != "z" {
		t.Fatalf("keys = %v, want [x z]", keys)
	}
}

func TestOrderedMapDeleteMissing(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Delete("missing") // should not panic
	if m.Len() != 1 {
		t.Fatalf("Len() = %d, want 1", m.Len())
	}
}

func TestOrderedMapLen(t *testing.T) {
	m := NewOrderedMap[int, string]()
	if m.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", m.Len())
	}
	m.Set(1, "a")
	m.Set(2, "b")
	if m.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", m.Len())
	}
}

func TestOrderedMapHas(t *testing.T) {
	m := NewOrderedMap[string, string]()
	m.Set("key", "val")
	if !m.Has("key") {
		t.Fatal("Has(key) should be true")
	}
	if m.Has("nope") {
		t.Fatal("Has(nope) should be false")
	}
}

func TestOrderedMapValues(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 10)
	m.Set("b", 20)
	m.Set("c", 30)

	vals := m.Values()
	want := []int{10, 20, 30}
	if len(vals) != len(want) {
		t.Fatalf("Values() len = %d, want %d", len(vals), len(want))
	}
	for i := range want {
		if vals[i] != want[i] {
			t.Fatalf("Values()[%d] = %d, want %d", i, vals[i], want[i])
		}
	}
}

func TestOrderedMapAll(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("x", 1)
	m.Set("y", 2)
	m.Set("z", 3)

	keys := make([]string, 0, 3)
	vals := make([]int, 0, 3)
	for k, v := range m.All() {
		keys = append(keys, k)
		vals = append(vals, v)
	}
	wantKeys := []string{"x", "y", "z"}
	wantVals := []int{1, 2, 3}
	for i := range wantKeys {
		if keys[i] != wantKeys[i] {
			t.Fatalf("key[%d] = %q, want %q", i, keys[i], wantKeys[i])
		}
		if vals[i] != wantVals[i] {
			t.Fatalf("val[%d] = %d, want %d", i, vals[i], wantVals[i])
		}
	}
}

func TestOrderedMapAllEarlyBreak(t *testing.T) {
	m := NewOrderedMap[string, int]()
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	count := 0
	for range m.All() {
		count++
		if count == 2 {
			break
		}
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
}

func TestOrderedMapEmpty(t *testing.T) {
	m := NewOrderedMap[string, string]()
	if m.Len() != 0 {
		t.Fatalf("Len() = %d, want 0", m.Len())
	}
	keys := m.Keys()
	if len(keys) != 0 {
		t.Fatalf("Keys() = %v, want empty", keys)
	}
	vals := m.Values()
	if len(vals) != 0 {
		t.Fatalf("Values() = %v, want empty", vals)
	}
	count := 0
	for range m.All() {
		count++
	}
	if count != 0 {
		t.Fatalf("All() yielded %d items, want 0", count)
	}
}

func BenchmarkOrderedMapSet(b *testing.B) {
	for b.Loop() {
		m := NewOrderedMap[int, int]()
		for i := range 100 {
			m.Set(i, i)
		}
	}
}

func BenchmarkOrderedMapGet(b *testing.B) {
	m := NewOrderedMap[int, int]()
	for i := range 100 {
		m.Set(i, i)
	}
	b.ResetTimer()
	for b.Loop() {
		m.Get(50)
	}
}

func BenchmarkOrderedMapAll(b *testing.B) {
	m := NewOrderedMap[int, int]()
	for i := range 100 {
		m.Set(i, i)
	}
	b.ResetTimer()
	for b.Loop() {
		for range m.All() {
		}
	}
}

func FuzzOrderedMapSetGet(f *testing.F) {
	f.Add("key", 42)
	f.Add("", 0)
	f.Add("hello world", -1)

	f.Fuzz(func(t *testing.T, k string, v int) {
		m := NewOrderedMap[string, int]()
		m.Set(k, v)
		got, ok := m.Get(k)
		if !ok {
			t.Fatal("key not found after Set")
		}
		if got != v {
			t.Fatalf("Get(%q) = %d, want %d", k, got, v)
		}
	})
}
