package maputil

import (
	"slices"
	"testing"
)

func TestKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(m)
	slices.Sort(keys)
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("Keys = %v, want [a b c]", keys)
	}
}

func TestKeys_Nil(t *testing.T) {
	keys := Keys[string, int](nil)
	if len(keys) != 0 {
		t.Errorf("Keys(nil) = %v, want empty", keys)
	}
}

func TestValues(t *testing.T) {
	m := map[string]int{"x": 10, "y": 20}
	vals := Values(m)
	slices.Sort(vals)
	if len(vals) != 2 || vals[0] != 10 || vals[1] != 20 {
		t.Errorf("Values = %v, want [10 20]", vals)
	}
}

func TestValues_Nil(t *testing.T) {
	vals := Values[string, int](nil)
	if len(vals) != 0 {
		t.Errorf("Values(nil) = %v, want empty", vals)
	}
}

func TestMerge(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"b": 3, "c": 4}
	got := Merge(a, b)
	want := map[string]int{"a": 1, "b": 3, "c": 4}
	if !Equal(got, want) {
		t.Errorf("Merge = %v, want %v", got, want)
	}
}

func TestMerge_Empty(t *testing.T) {
	got := Merge[string, int]()
	if len(got) != 0 {
		t.Errorf("Merge() = %v, want empty", got)
	}
}

func TestMerge_Nil(t *testing.T) {
	got := Merge[string, int](nil, map[string]int{"a": 1})
	want := map[string]int{"a": 1}
	if !Equal(got, want) {
		t.Errorf("Merge(nil, m) = %v, want %v", got, want)
	}
}

func TestPick(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := Pick(m, "a", "c", "missing")
	want := map[string]int{"a": 1, "c": 3}
	if !Equal(got, want) {
		t.Errorf("Pick = %v, want %v", got, want)
	}
}

func TestPick_Empty(t *testing.T) {
	got := Pick(map[string]int{"a": 1})
	if len(got) != 0 {
		t.Errorf("Pick() = %v, want empty", got)
	}
}

func TestOmit(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	got := Omit(m, "b")
	want := map[string]int{"a": 1, "c": 3}
	if !Equal(got, want) {
		t.Errorf("Omit = %v, want %v", got, want)
	}
}

func TestOmit_NonexistentKey(t *testing.T) {
	m := map[string]int{"a": 1}
	got := Omit(m, "z")
	want := map[string]int{"a": 1}
	if !Equal(got, want) {
		t.Errorf("Omit = %v, want %v", got, want)
	}
}

func TestFilter(t *testing.T) {
	m := map[string]int{"a": 1, "b": 5, "c": 3}
	got := Filter(m, func(_ string, v int) bool { return v > 2 })
	want := map[string]int{"b": 5, "c": 3}
	if !Equal(got, want) {
		t.Errorf("Filter = %v, want %v", got, want)
	}
}

func TestFilter_NoMatch(t *testing.T) {
	m := map[string]int{"a": 1}
	got := Filter(m, func(_ string, _ int) bool { return false })
	if len(got) != 0 {
		t.Errorf("Filter = %v, want empty", got)
	}
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	got := MapValues(m, func(v int) int { return v * 10 })
	want := map[string]int{"a": 10, "b": 20}
	if !Equal(got, want) {
		t.Errorf("MapValues = %v, want %v", got, want)
	}
}

func TestMapValues_TypeChange(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	got := MapValues(m, func(v int) string {
		if v == 1 {
			return "one"
		}
		return "two"
	})
	want := map[string]string{"a": "one", "b": "two"}
	if !Equal(got, want) {
		t.Errorf("MapValues = %v, want %v", got, want)
	}
}

func TestInvert(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2}
	got := Invert(m)
	want := map[int]string{1: "a", 2: "b"}
	if !Equal(got, want) {
		t.Errorf("Invert = %v, want %v", got, want)
	}
}

func TestInvert_Empty(t *testing.T) {
	got := Invert(map[string]int{})
	if len(got) != 0 {
		t.Errorf("Invert({}) = %v, want empty", got)
	}
}

func TestEqual_Same(t *testing.T) {
	a := map[string]int{"a": 1, "b": 2}
	b := map[string]int{"a": 1, "b": 2}
	if !Equal(a, b) {
		t.Error("Equal should be true for identical maps")
	}
}

func TestEqual_Different(t *testing.T) {
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 2}
	if Equal(a, b) {
		t.Error("Equal should be false for different values")
	}
}

func TestEqual_DifferentKeys(t *testing.T) {
	a := map[string]int{"a": 1}
	b := map[string]int{"b": 1}
	if Equal(a, b) {
		t.Error("Equal should be false for different keys")
	}
}

func TestEqual_DifferentLengths(t *testing.T) {
	a := map[string]int{"a": 1}
	b := map[string]int{"a": 1, "b": 2}
	if Equal(a, b) {
		t.Error("Equal should be false for different lengths")
	}
}

func TestEqual_BothNil(t *testing.T) {
	if !Equal[string, int](nil, nil) {
		t.Error("Equal should be true for two nil maps")
	}
}

func BenchmarkKeys(b *testing.B) {
	m := make(map[int]int, 1000)
	for i := range 1000 {
		m[i] = i
	}
	for b.Loop() {
		Keys(m)
	}
}

func BenchmarkMerge(b *testing.B) {
	a := make(map[int]int, 100)
	c := make(map[int]int, 100)
	for i := range 100 {
		a[i] = i
		c[i+50] = i
	}
	for b.Loop() {
		Merge(a, c)
	}
}

func BenchmarkFilter(b *testing.B) {
	m := make(map[int]int, 1000)
	for i := range 1000 {
		m[i] = i
	}
	for b.Loop() {
		Filter(m, func(_ int, v int) bool { return v%2 == 0 })
	}
}
