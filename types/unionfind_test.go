package types

import (
	"testing"
)

func TestNewUnionFind(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)
	if uf.Count() != 5 {
		t.Fatalf("Count() = %d, want 5", uf.Count())
	}
	if uf.Size() != 5 {
		t.Fatalf("Size() = %d, want 5", uf.Size())
	}
}

func TestNewUnionFindNegative(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(-1)
	if uf.Count() != 0 {
		t.Fatalf("Count() = %d, want 0", uf.Count())
	}
	if uf.Size() != 0 {
		t.Fatalf("Size() = %d, want 0", uf.Size())
	}
}

func TestNewUnionFindZero(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(0)
	if uf.Count() != 0 {
		t.Fatalf("Count() = %d, want 0", uf.Count())
	}
}

func TestFind(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)
	for i := range 5 {
		if got := uf.Find(i); got != i {
			t.Fatalf("Find(%d) = %d, want %d", i, got, i)
		}
	}
}

func TestFindOutOfRange(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(3)
	if got := uf.Find(-1); got != -1 {
		t.Fatalf("Find(-1) = %d, want -1", got)
	}
	if got := uf.Find(3); got != -1 {
		t.Fatalf("Find(3) = %d, want -1", got)
	}
}

func TestUnion(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)

	if !uf.Union(0, 1) {
		t.Fatal("Union(0,1) should return true")
	}
	if uf.Count() != 4 {
		t.Fatalf("Count() = %d, want 4", uf.Count())
	}
	if !uf.Connected(0, 1) {
		t.Fatal("0 and 1 should be connected")
	}
}

func TestUnionSameSet(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)
	uf.Union(0, 1)
	if uf.Union(0, 1) {
		t.Fatal("Union of already connected elements should return false")
	}
}

func TestUnionOutOfRange(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(3)
	if uf.Union(-1, 0) {
		t.Fatal("Union with out-of-range should return false")
	}
	if uf.Union(0, 5) {
		t.Fatal("Union with out-of-range should return false")
	}
}

func TestConnected(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)
	if uf.Connected(0, 1) {
		t.Fatal("0 and 1 should not be connected initially")
	}
	uf.Union(0, 1)
	if !uf.Connected(0, 1) {
		t.Fatal("0 and 1 should be connected after union")
	}
}

func TestConnectedOutOfRange(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(3)
	if uf.Connected(-1, 0) {
		t.Fatal("Connected with out-of-range should be false")
	}
	if uf.Connected(0, 5) {
		t.Fatal("Connected with out-of-range should be false")
	}
}

func TestUnionByRank(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(6)
	// Build a tree of rank 1: {0,1}
	uf.Union(0, 1)
	// Build a tree of rank 1: {2,3}
	uf.Union(2, 3)
	// Merge two rank-1 trees → rank 2
	uf.Union(0, 2)
	// {4} has rank 0, merging with rank-2 tree
	uf.Union(0, 4)

	if uf.Count() != 2 {
		t.Fatalf("Count() = %d, want 2", uf.Count())
	}
	// All should be connected
	for _, x := range []int{0, 1, 2, 3, 4} {
		if !uf.Connected(0, x) {
			t.Fatalf("0 and %d should be connected", x)
		}
	}
	// 5 is separate
	if uf.Connected(0, 5) {
		t.Fatal("0 and 5 should not be connected")
	}
}

func TestPathCompression(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(10)
	// Chain: 0-1-2-3-4
	for i := range 4 {
		uf.Union(i, i+1)
	}
	// Find should compress paths
	root := uf.Find(4)
	// After compression, 4's parent should be closer to root
	if uf.Find(4) != root {
		t.Fatal("Find should return consistent root")
	}
}

func TestTransitiveConnectivity(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(5)
	uf.Union(0, 1)
	uf.Union(1, 2)
	uf.Union(3, 4)

	if !uf.Connected(0, 2) {
		t.Fatal("transitive: 0 and 2 should be connected")
	}
	if uf.Connected(0, 3) {
		t.Fatal("0 and 3 should not be connected")
	}

	uf.Union(2, 3)
	if !uf.Connected(0, 4) {
		t.Fatal("after bridge: 0 and 4 should be connected")
	}
	if uf.Count() != 1 {
		t.Fatalf("Count() = %d, want 1", uf.Count())
	}
}

func BenchmarkUnionFind(b *testing.B) {
	for b.Loop() {
		uf := NewUnionFind(1000)
		for i := range 999 {
			uf.Union(i, i+1)
		}
	}
}

func BenchmarkFind(b *testing.B) {
	uf := NewUnionFind(1000)
	for i := range 999 {
		uf.Union(i, i+1)
	}
	b.ResetTimer()
	for b.Loop() {
		uf.Find(999)
	}
}

func TestUnionRankSwap(t *testing.T) {
	t.Parallel()
	uf := NewUnionFind(4)
	// Build rank-1 tree: {0, 1}
	uf.Union(0, 1)
	// {2} has rank 0. Union(2, 0) means rx=Find(2)=2 (rank 0), ry=Find(0)=root (rank 1).
	// This triggers the swap since rank[rx] < rank[ry].
	if !uf.Union(2, 0) {
		t.Fatal("Union(2, 0) should return true")
	}
	if !uf.Connected(2, 1) {
		t.Fatal("2 and 1 should be connected")
	}
}
