package types

import "testing"

func TestNewInterval(t *testing.T) {
	t.Parallel()
	iv := NewInterval(1, 10)
	if iv.Lo != 1 || iv.Hi != 10 {
		t.Errorf("got {%d, %d}; want {1, 10}", iv.Lo, iv.Hi)
	}
}

func TestNewIntervalSwap(t *testing.T) {
	t.Parallel()
	iv := NewInterval(10, 1)
	if iv.Lo != 1 || iv.Hi != 10 {
		t.Errorf("got {%d, %d}; want {1, 10}", iv.Lo, iv.Hi)
	}
}

func TestIntervalContains(t *testing.T) {
	t.Parallel()
	iv := NewInterval(1, 10)
	tests := []struct {
		v    int
		want bool
	}{
		{0, false},
		{1, true},
		{5, true},
		{10, true},
		{11, false},
	}
	for _, tt := range tests {
		if got := iv.Contains(tt.v); got != tt.want {
			t.Errorf("Contains(%d) = %v; want %v", tt.v, got, tt.want)
		}
	}
}

func TestIntervalOverlaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b Interval[int]
		want bool
	}{
		{"overlap", NewInterval(1, 5), NewInterval(3, 8), true},
		{"touch", NewInterval(1, 5), NewInterval(5, 10), true},
		{"disjoint", NewInterval(1, 3), NewInterval(5, 8), false},
		{"contained", NewInterval(1, 10), NewInterval(3, 5), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.a.Overlaps(tt.b); got != tt.want {
				t.Errorf("Overlaps = %v; want %v", got, tt.want)
			}
		})
	}
}

func TestIntervalMerge(t *testing.T) {
	t.Parallel()
	a := NewInterval(1, 5)
	b := NewInterval(3, 10)
	m := a.Merge(b)
	if m.Lo != 1 || m.Hi != 10 {
		t.Errorf("Merge = {%d, %d}; want {1, 10}", m.Lo, m.Hi)
	}
}

func TestIntervalMergeDisjoint(t *testing.T) {
	t.Parallel()
	a := NewInterval(1, 3)
	b := NewInterval(7, 10)
	m := a.Merge(b)
	if m.Lo != 1 || m.Hi != 10 {
		t.Errorf("Merge = {%d, %d}; want {1, 10}", m.Lo, m.Hi)
	}
}

func TestIntervalIntersect(t *testing.T) {
	t.Parallel()
	a := NewInterval(1, 5)
	b := NewInterval(3, 8)
	inter, ok := a.Intersect(b)
	if !ok {
		t.Fatal("Intersect returned false")
	}
	if inter.Lo != 3 || inter.Hi != 5 {
		t.Errorf("Intersect = {%d, %d}; want {3, 5}", inter.Lo, inter.Hi)
	}
}

func TestIntervalIntersectNone(t *testing.T) {
	t.Parallel()
	a := NewInterval(1, 3)
	b := NewInterval(5, 8)
	_, ok := a.Intersect(b)
	if ok {
		t.Error("Intersect should return false for disjoint")
	}
}

func TestIntervalEmpty(t *testing.T) {
	t.Parallel()
	var iv Interval[int] // zero value
	if iv.Empty() {
		t.Error("zero Interval[int] should not be empty (0 == 0)")
	}
}

func TestIntervalEqual(t *testing.T) {
	t.Parallel()
	a := NewInterval(1, 10)
	b := NewInterval(1, 10)
	c := NewInterval(2, 10)
	if !a.Equal(b) {
		t.Error("Equal should be true for same intervals")
	}
	if a.Equal(c) {
		t.Error("Equal should be false for different intervals")
	}
}

func TestIntervalClamp(t *testing.T) {
	t.Parallel()
	iv := NewInterval(0, 100)
	tests := []struct {
		v    int
		want int
	}{
		{-5, 0},
		{0, 0},
		{50, 50},
		{100, 100},
		{200, 100},
	}
	for _, tt := range tests {
		if got := iv.Clamp(tt.v); got != tt.want {
			t.Errorf("Clamp(%d) = %d; want %d", tt.v, got, tt.want)
		}
	}
}

func TestIntervalString(t *testing.T) {
	t.Parallel()
	iv := NewInterval("apple", "cherry")
	if !iv.Contains("banana") {
		t.Error("Contains(banana) = false; want true")
	}
	if iv.Contains("date") {
		t.Error("Contains(date) = true; want false")
	}
}

func BenchmarkIntervalContains(b *testing.B) {
	iv := NewInterval(0, 1000)
	for range b.N {
		iv.Contains(500)
	}
}

func BenchmarkIntervalOverlaps(b *testing.B) {
	a := NewInterval(0, 100)
	other := NewInterval(50, 150)
	for range b.N {
		a.Overlaps(other)
	}
}

func FuzzIntervalContains(f *testing.F) {
	f.Add(0, 100, 50)
	f.Add(-10, 10, 0)
	f.Fuzz(func(t *testing.T, lo, hi, v int) {
		iv := NewInterval(lo, hi)
		got := iv.Contains(v)
		// Verify manually
		realLo, realHi := lo, hi
		if lo > hi {
			realLo, realHi = hi, lo
		}
		want := v >= realLo && v <= realHi
		if got != want {
			t.Errorf("Contains(%d) in [%d,%d] = %v; want %v", v, realLo, realHi, got, want)
		}
	})
}

func TestIntervalMergeLowerOther(t *testing.T) {
t.Parallel()
// other.Lo < iv.Lo
a := NewInterval(5, 10)
b := NewInterval(1, 7)
m := a.Merge(b)
if m.Lo != 1 || m.Hi != 10 {
t.Errorf("Merge = {%d, %d}; want {1, 10}", m.Lo, m.Hi)
}
}

func TestIntervalIntersectContained(t *testing.T) {
t.Parallel()
// other is fully contained in iv; other.Hi < iv.Hi
a := NewInterval(1, 10)
b := NewInterval(3, 7)
inter, ok := a.Intersect(b)
if !ok {
t.Fatal("expected overlap")
}
if inter.Lo != 3 || inter.Hi != 7 {
t.Errorf("Intersect = {%d, %d}; want {3, 7}", inter.Lo, inter.Hi)
}
}
