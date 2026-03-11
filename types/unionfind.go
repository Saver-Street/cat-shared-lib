package types

// UnionFind implements a disjoint-set / union–find data structure with
// union by rank and path compression for near-O(1) amortized operations.
type UnionFind struct {
	parent []int
	rank   []int
	count  int // number of disjoint sets
}

// NewUnionFind creates a UnionFind with n elements (0 to n-1), each in its own set.
func NewUnionFind(n int) *UnionFind {
	if n < 0 {
		n = 0
	}
	parent := make([]int, n)
	rank := make([]int, n)
	for i := range parent {
		parent[i] = i
	}
	return &UnionFind{parent: parent, rank: rank, count: n}
}

// Find returns the representative (root) of the set containing x.
// It uses path compression so that subsequent lookups are faster.
// Returns -1 if x is out of range.
func (uf *UnionFind) Find(x int) int {
	if x < 0 || x >= len(uf.parent) {
		return -1
	}
	for uf.parent[x] != x {
		uf.parent[x] = uf.parent[uf.parent[x]] // path halving
		x = uf.parent[x]
	}
	return x
}

// Union merges the sets containing x and y. Returns true if the sets
// were disjoint and have been merged, false if they were already in the
// same set or either index is out of range.
func (uf *UnionFind) Union(x, y int) bool {
	rx, ry := uf.Find(x), uf.Find(y)
	if rx < 0 || ry < 0 || rx == ry {
		return false
	}
	if uf.rank[rx] < uf.rank[ry] {
		rx, ry = ry, rx
	}
	uf.parent[ry] = rx
	if uf.rank[rx] == uf.rank[ry] {
		uf.rank[rx]++
	}
	uf.count--
	return true
}

// Connected reports whether x and y are in the same set.
func (uf *UnionFind) Connected(x, y int) bool {
	rx, ry := uf.Find(x), uf.Find(y)
	return rx >= 0 && ry >= 0 && rx == ry
}

// Count returns the number of disjoint sets.
func (uf *UnionFind) Count() int {
	return uf.count
}

// Size returns the total number of elements.
func (uf *UnionFind) Size() int {
	return len(uf.parent)
}
