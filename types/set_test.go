package types

import (
	"sort"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNewSet(t *testing.T) {
	s := NewSet(1, 2, 3)
	testkit.AssertEqual(t, s.Len(), 3)
	testkit.AssertTrue(t, s.Contains(1))
	testkit.AssertTrue(t, s.Contains(2))
	testkit.AssertTrue(t, s.Contains(3))
}

func TestNewSet_Empty(t *testing.T) {
	s := NewSet[string]()
	testkit.AssertEqual(t, s.Len(), 0)
}

func TestNewSet_Dedup(t *testing.T) {
	s := NewSet(1, 1, 2, 2, 3)
	testkit.AssertEqual(t, s.Len(), 3)
}

func TestSet_Add(t *testing.T) {
	s := NewSet[int]()
	s.Add(1, 2)
	testkit.AssertEqual(t, s.Len(), 2)
	s.Add(2, 3)
	testkit.AssertEqual(t, s.Len(), 3)
}

func TestSet_AddToZeroValue(t *testing.T) {
	var s Set[string]
	s.Add("a")
	testkit.AssertEqual(t, s.Len(), 1)
	testkit.AssertTrue(t, s.Contains("a"))
}

func TestSet_Remove(t *testing.T) {
	s := NewSet(1, 2, 3)
	s.Remove(2)
	testkit.AssertEqual(t, s.Len(), 2)
	testkit.AssertTrue(t, !s.Contains(2))
}

func TestSet_RemoveNonExistent(t *testing.T) {
	s := NewSet(1, 2)
	s.Remove(99)
	testkit.AssertEqual(t, s.Len(), 2)
}

func TestSet_Contains(t *testing.T) {
	s := NewSet("a", "b")
	testkit.AssertTrue(t, s.Contains("a"))
	testkit.AssertTrue(t, !s.Contains("c"))
}

func TestSet_Values(t *testing.T) {
	s := NewSet(3, 1, 2)
	vals := s.Values()
	testkit.AssertEqual(t, len(vals), 3)
	sort.Ints(vals)
	testkit.AssertEqual(t, vals[0], 1)
	testkit.AssertEqual(t, vals[1], 2)
	testkit.AssertEqual(t, vals[2], 3)
}

func TestSet_Union(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(3, 4, 5)
	u := a.Union(b)
	testkit.AssertEqual(t, u.Len(), 5)
}

func TestSet_Intersect(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)
	i := a.Intersect(b)
	testkit.AssertEqual(t, i.Len(), 2)
	testkit.AssertTrue(t, i.Contains(2))
	testkit.AssertTrue(t, i.Contains(3))
}

func TestSet_Intersect_Empty(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(3, 4)
	i := a.Intersect(b)
	testkit.AssertEqual(t, i.Len(), 0)
}

func TestSet_Intersect_Asymmetric(t *testing.T) {
	// Ensure the smaller-set optimization works both ways.
	big := NewSet(1, 2, 3, 4, 5)
	small := NewSet(3)
	i := big.Intersect(small)
	testkit.AssertEqual(t, i.Len(), 1)
	testkit.AssertTrue(t, i.Contains(3))
}

func TestSet_Diff(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(2, 3, 4)
	d := a.Diff(b)
	testkit.AssertEqual(t, d.Len(), 1)
	testkit.AssertTrue(t, d.Contains(1))
}

func TestSet_Diff_Empty(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(1, 2, 3)
	d := a.Diff(b)
	testkit.AssertEqual(t, d.Len(), 0)
}

func TestSet_Equal(t *testing.T) {
	a := NewSet(1, 2, 3)
	b := NewSet(3, 2, 1)
	testkit.AssertTrue(t, a.Equal(b))
}

func TestSet_Equal_DifferentLen(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(1, 2, 3)
	testkit.AssertTrue(t, !a.Equal(b))
}

func TestSet_Equal_SameLenDifferent(t *testing.T) {
	a := NewSet(1, 2)
	b := NewSet(1, 3)
	testkit.AssertTrue(t, !a.Equal(b))
}

func TestSet_StringType(t *testing.T) {
	s := NewSet("hello", "world")
	testkit.AssertEqual(t, s.Len(), 2)
	testkit.AssertTrue(t, s.Contains("hello"))
}

func BenchmarkSet_Add(b *testing.B) {
	for b.Loop() {
		s := NewSet[int]()
		for i := range 100 {
			s.Add(i)
		}
	}
}

func BenchmarkSet_Contains(b *testing.B) {
	s := NewSet[int]()
	for i := range 1000 {
		s.Add(i)
	}
	for b.Loop() {
		s.Contains(500)
	}
}

func BenchmarkSet_Union(b *testing.B) {
	a := NewSet[int]()
	c := NewSet[int]()
	for i := range 100 {
		a.Add(i)
		c.Add(i + 50)
	}
	for b.Loop() {
		a.Union(c)
	}
}
