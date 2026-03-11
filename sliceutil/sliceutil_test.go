package sliceutil

import (
	"strings"
	"testing"
)

// ─── Partition ───────────────────────────────────────────────────────────────

func TestPartition_SplitEvenOdd(t *testing.T) {
	even, odd := Partition([]int{1, 2, 3, 4, 5}, func(n int) bool { return n%2 == 0 })
	if len(even) != 2 || len(odd) != 3 {
		t.Fatalf("got even=%v odd=%v", even, odd)
	}
}

func TestPartition_Empty(t *testing.T) {
	matched, unmatched := Partition([]int{}, func(int) bool { return true })
	if len(matched) != 0 || len(unmatched) != 0 {
		t.Fatal("expected empty slices")
	}
}

func TestPartition_Nil(t *testing.T) {
	matched, unmatched := Partition[int](nil, func(int) bool { return true })
	if len(matched) != 0 || len(unmatched) != 0 {
		t.Fatal("expected empty slices for nil input")
	}
}

func TestPartition_AllMatch(t *testing.T) {
	m, u := Partition([]int{2, 4, 6}, func(n int) bool { return n%2 == 0 })
	if len(m) != 3 || len(u) != 0 {
		t.Fatalf("got matched=%v unmatched=%v", m, u)
	}
}

// ─── Reduce ──────────────────────────────────────────────────────────────────

func TestReduce_Sum(t *testing.T) {
	sum := Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int { return acc + n })
	if sum != 10 {
		t.Fatalf("got %d, want 10", sum)
	}
}

func TestReduce_Empty(t *testing.T) {
	result := Reduce([]int{}, 42, func(acc, n int) int { return acc + n })
	if result != 42 {
		t.Fatalf("got %d, want 42", result)
	}
}

func TestReduce_StringConcat(t *testing.T) {
	result := Reduce([]string{"a", "b", "c"}, "", func(acc, s string) string { return acc + s })
	if result != "abc" {
		t.Fatalf("got %q, want %q", result, "abc")
	}
}

// ─── Any ─────────────────────────────────────────────────────────────────────

func TestAny_True(t *testing.T) {
	if !Any([]int{1, 2, 3}, func(n int) bool { return n > 2 }) {
		t.Fatal("expected true")
	}
}

func TestAny_False(t *testing.T) {
	if Any([]int{1, 2, 3}, func(n int) bool { return n > 5 }) {
		t.Fatal("expected false")
	}
}

func TestAny_Empty(t *testing.T) {
	if Any([]int{}, func(int) bool { return true }) {
		t.Fatal("expected false for empty")
	}
}

// ─── All ─────────────────────────────────────────────────────────────────────

func TestAll_True(t *testing.T) {
	if !All([]int{2, 4, 6}, func(n int) bool { return n%2 == 0 }) {
		t.Fatal("expected true")
	}
}

func TestAll_False(t *testing.T) {
	if All([]int{2, 3, 6}, func(n int) bool { return n%2 == 0 }) {
		t.Fatal("expected false")
	}
}

func TestAll_Empty(t *testing.T) {
	if !All([]int{}, func(int) bool { return false }) {
		t.Fatal("expected true for empty")
	}
}

// ─── None ────────────────────────────────────────────────────────────────────

func TestNone_True(t *testing.T) {
	if !None([]int{1, 3, 5}, func(n int) bool { return n%2 == 0 }) {
		t.Fatal("expected true")
	}
}

func TestNone_False(t *testing.T) {
	if None([]int{1, 2, 3}, func(n int) bool { return n%2 == 0 }) {
		t.Fatal("expected false")
	}
}

// ─── Find ────────────────────────────────────────────────────────────────────

func TestFind_Found(t *testing.T) {
	v, ok := Find([]string{"apple", "banana", "cherry"}, func(s string) bool { return strings.HasPrefix(s, "b") })
	if !ok || v != "banana" {
		t.Fatalf("got %q, %v", v, ok)
	}
}

func TestFind_NotFound(t *testing.T) {
	_, ok := Find([]int{1, 2, 3}, func(n int) bool { return n > 10 })
	if ok {
		t.Fatal("expected not found")
	}
}

func TestFind_Empty(t *testing.T) {
	_, ok := Find([]int{}, func(int) bool { return true })
	if ok {
		t.Fatal("expected not found for empty")
	}
}

// ─── FindIndex ───────────────────────────────────────────────────────────────

func TestFindIndex_Found(t *testing.T) {
	idx := FindIndex([]int{10, 20, 30}, func(n int) bool { return n == 20 })
	if idx != 1 {
		t.Fatalf("got %d, want 1", idx)
	}
}

func TestFindIndex_NotFound(t *testing.T) {
	idx := FindIndex([]int{10, 20, 30}, func(n int) bool { return n == 99 })
	if idx != -1 {
		t.Fatalf("got %d, want -1", idx)
	}
}

// ─── Last ────────────────────────────────────────────────────────────────────

func TestLast_NonEmpty(t *testing.T) {
	v, ok := Last([]int{1, 2, 3})
	if !ok || v != 3 {
		t.Fatalf("got %d, %v", v, ok)
	}
}

func TestLast_Empty(t *testing.T) {
	_, ok := Last([]int{})
	if ok {
		t.Fatal("expected false for empty")
	}
}

func TestLast_SingleElement(t *testing.T) {
	v, ok := Last([]string{"only"})
	if !ok || v != "only" {
		t.Fatalf("got %q, %v", v, ok)
	}
}

// ─── Take ────────────────────────────────────────────────────────────────────

func TestTake_Normal(t *testing.T) {
	result := Take([]int{1, 2, 3, 4, 5}, 3)
	if len(result) != 3 || result[2] != 3 {
		t.Fatalf("got %v", result)
	}
}

func TestTake_MoreThanLen(t *testing.T) {
	result := Take([]int{1, 2}, 5)
	if len(result) != 2 {
		t.Fatalf("got %v", result)
	}
}

func TestTake_Negative(t *testing.T) {
	result := Take([]int{1, 2, 3}, -1)
	if len(result) != 0 {
		t.Fatalf("got %v", result)
	}
}

// ─── Drop ────────────────────────────────────────────────────────────────────

func TestDrop_Normal(t *testing.T) {
	result := Drop([]int{1, 2, 3, 4, 5}, 2)
	if len(result) != 3 || result[0] != 3 {
		t.Fatalf("got %v", result)
	}
}

func TestDrop_MoreThanLen(t *testing.T) {
	result := Drop([]int{1, 2}, 5)
	if len(result) != 0 {
		t.Fatalf("got %v", result)
	}
}

func TestDrop_Negative(t *testing.T) {
	result := Drop([]int{1, 2, 3}, -1)
	if len(result) != 3 {
		t.Fatalf("got %v", result)
	}
}

// ─── Associate ───────────────────────────────────────────────────────────────

func TestAssociate_Normal(t *testing.T) {
	type user struct {
		ID   int
		Name string
	}
	users := []user{{1, "Alice"}, {2, "Bob"}}
	m := Associate(users, func(u user) int { return u.ID })
	if m[1].Name != "Alice" || m[2].Name != "Bob" {
		t.Fatalf("got %v", m)
	}
}

func TestAssociate_Empty(t *testing.T) {
	m := Associate([]string{}, func(s string) string { return s })
	if len(m) != 0 {
		t.Fatal("expected empty map")
	}
}

func TestAssociate_DuplicateKey(t *testing.T) {
	m := Associate([]string{"a", "ab", "abc"}, func(s string) int { return len(s) })
	if len(m) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(m))
	}
}

// ─── FlatMap ─────────────────────────────────────────────────────────────────

func TestFlatMap_Normal(t *testing.T) {
	result := FlatMap([]string{"hello", "world"}, func(s string) []byte { return []byte(s) })
	if len(result) != 10 {
		t.Fatalf("got len %d", len(result))
	}
}

func TestFlatMap_Empty(t *testing.T) {
	result := FlatMap([]int{}, func(int) []int { return []int{1} })
	if len(result) != 0 {
		t.Fatalf("got %v", result)
	}
}

// ─── Count ───────────────────────────────────────────────────────────────────

func TestCount_Normal(t *testing.T) {
	n := Count([]int{1, 2, 3, 4, 5}, func(v int) bool { return v > 3 })
	if n != 2 {
		t.Fatalf("got %d, want 2", n)
	}
}

func TestCount_Empty(t *testing.T) {
	n := Count([]int{}, func(int) bool { return true })
	if n != 0 {
		t.Fatalf("got %d, want 0", n)
	}
}

// ─── ForEach ─────────────────────────────────────────────────────────────────

func TestForEach_Normal(t *testing.T) {
	var indices []int
	ForEach([]string{"a", "b", "c"}, func(i int, _ string) { indices = append(indices, i) })
	if len(indices) != 3 || indices[2] != 2 {
		t.Fatalf("got %v", indices)
	}
}

func TestForEach_Empty(t *testing.T) {
	called := false
	ForEach([]int{}, func(int, int) { called = true })
	if called {
		t.Fatal("should not call fn on empty slice")
	}
}

// ─── Benchmarks ──────────────────────────────────────────────────────────────

func BenchmarkReduce(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		Reduce(items, 0, func(acc, n int) int { return acc + n })
	}
}

func BenchmarkFind(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		Find(items, func(n int) bool { return n == 999 })
	}
}

func BenchmarkAssociate(b *testing.B) {
	items := make([]int, 1000)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		Associate(items, func(n int) int { return n })
	}
}

// ─── Fuzz ────────────────────────────────────────────────────────────────────

func FuzzPartition(f *testing.F) {
	f.Add(5)
	f.Fuzz(func(t *testing.T, n int) {
		if n < 0 {
			n = -n
		}
		if n > 10000 {
			n = 10000
		}
		items := make([]int, n)
		for i := range items {
			items[i] = i
		}
		matched, unmatched := Partition(items, func(v int) bool { return v%2 == 0 })
		if len(matched)+len(unmatched) != len(items) {
			t.Fatalf("len(matched)=%d + len(unmatched)=%d != len(items)=%d",
				len(matched), len(unmatched), len(items))
		}
	})
}
