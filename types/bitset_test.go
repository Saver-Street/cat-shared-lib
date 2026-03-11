package types

import (
	"testing"
)

func TestBitSetSetAndTest(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(100)
	bs.Set(0)
	bs.Set(42)
	bs.Set(99)

	if !bs.Test(0) {
		t.Error("bit 0 should be set")
	}
	if !bs.Test(42) {
		t.Error("bit 42 should be set")
	}
	if !bs.Test(99) {
		t.Error("bit 99 should be set")
	}
	if bs.Test(1) {
		t.Error("bit 1 should not be set")
	}
	if bs.Test(200) {
		t.Error("bit 200 should not be set (out of range)")
	}
}

func TestBitSetClear(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(64)
	bs.Set(5)
	bs.Clear(5)
	if bs.Test(5) {
		t.Error("bit 5 should be cleared")
	}
	if bs.Len() != 0 {
		t.Errorf("Len() = %d; want 0", bs.Len())
	}
	// Clear non-existent bit should not panic
	bs.Clear(999)
}

func TestBitSetToggle(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(64)
	bs.Toggle(3)
	if !bs.Test(3) {
		t.Error("bit 3 should be set after toggle")
	}
	bs.Toggle(3)
	if bs.Test(3) {
		t.Error("bit 3 should be cleared after second toggle")
	}
}

func TestBitSetLen(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(64)
	if bs.Len() != 0 {
		t.Errorf("Len() = %d; want 0", bs.Len())
	}
	bs.Set(1)
	bs.Set(2)
	bs.Set(3)
	if bs.Len() != 3 {
		t.Errorf("Len() = %d; want 3", bs.Len())
	}
	// Setting same bit again should not change count
	bs.Set(2)
	if bs.Len() != 3 {
		t.Errorf("Len() = %d after re-set; want 3", bs.Len())
	}
}

func TestBitSetCap(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(100)
	if bs.Cap() < 100 {
		t.Errorf("Cap() = %d; want >= 100", bs.Cap())
	}
}

func TestBitSetGrow(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(8)
	bs.Set(200)
	if !bs.Test(200) {
		t.Error("bit 200 should be set after grow")
	}
}

func TestBitSetString(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(64)
	bs.Set(0)
	bs.Set(3)
	bs.Set(7)
	want := "{0, 3, 7}"
	got := bs.String()
	if got != want {
		t.Errorf("String() = %q; want %q", got, want)
	}
}

func TestBitSetStringEmpty(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(64)
	if bs.String() != "{}" {
		t.Errorf("String() = %q; want {}", bs.String())
	}
}

func TestBitSetUnion(t *testing.T) {
	t.Parallel()
	a := NewBitSet(64)
	a.Set(1)
	a.Set(3)
	b := NewBitSet(128)
	b.Set(3)
	b.Set(100)

	u := a.Union(b)
	if !u.Test(1) || !u.Test(3) || !u.Test(100) {
		t.Error("union should contain all bits from both sets")
	}
	if u.Len() != 3 {
		t.Errorf("union Len() = %d; want 3", u.Len())
	}
}

func TestBitSetIntersect(t *testing.T) {
	t.Parallel()
	a := NewBitSet(64)
	a.Set(1)
	a.Set(3)
	a.Set(5)
	b := NewBitSet(64)
	b.Set(3)
	b.Set(5)
	b.Set(7)

	i := a.Intersect(b)
	if !i.Test(3) || !i.Test(5) {
		t.Error("intersect should contain common bits")
	}
	if i.Test(1) || i.Test(7) {
		t.Error("intersect should not contain non-common bits")
	}
	if i.Len() != 2 {
		t.Errorf("intersect Len() = %d; want 2", i.Len())
	}
}

func TestBitSetDiff(t *testing.T) {
	t.Parallel()
	a := NewBitSet(64)
	a.Set(1)
	a.Set(3)
	a.Set(5)
	b := NewBitSet(64)
	b.Set(3)

	d := a.Diff(b)
	if !d.Test(1) || !d.Test(5) {
		t.Error("diff should contain bits in a but not b")
	}
	if d.Test(3) {
		t.Error("diff should not contain bits in both sets")
	}
	if d.Len() != 2 {
		t.Errorf("diff Len() = %d; want 2", d.Len())
	}
}

func TestBitSetIntersectDifferentSizes(t *testing.T) {
	t.Parallel()
	a := NewBitSet(128)
	a.Set(1)
	a.Set(100)
	b := NewBitSet(32)
	b.Set(1)

	i := a.Intersect(b)
	if !i.Test(1) {
		t.Error("intersect should find bit 1")
	}
	if i.Len() != 1 {
		t.Errorf("intersect Len() = %d; want 1", i.Len())
	}
}

func TestNewBitSetZeroSize(t *testing.T) {
	t.Parallel()
	bs := NewBitSet(0)
	bs.Set(5)
	if !bs.Test(5) {
		t.Error("bit 5 should be set even with size 0 init")
	}
}

func BenchmarkBitSetSet(b *testing.B) {
	bs := NewBitSet(b.N)
	for i := range b.N {
		bs.Set(i)
	}
}

func BenchmarkBitSetTest(b *testing.B) {
	bs := NewBitSet(1000)
	for i := range 1000 {
		bs.Set(i)
	}
	b.ResetTimer()
	for i := range b.N {
		bs.Test(i % 1000)
	}
}

func BenchmarkBitSetUnion(b *testing.B) {
	a := NewBitSet(1000)
	c := NewBitSet(1000)
	for i := range 500 {
		a.Set(i)
		c.Set(i + 250)
	}
	b.ResetTimer()
	for range b.N {
		a.Union(c)
	}
}

func FuzzBitSetSetTest(f *testing.F) {
	f.Add(0)
	f.Add(63)
	f.Add(64)
	f.Add(1000)
	f.Fuzz(func(t *testing.T, i int) {
		if i < 0 {
			return
		}
		if i > 100000 {
			return
		}
		bs := NewBitSet(0)
		bs.Set(i)
		if !bs.Test(i) {
			t.Errorf("bit %d should be set", i)
		}
	})
}
