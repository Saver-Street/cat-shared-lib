package types

import (
	"sort"
	"testing"
)

func TestGraphNew(t *testing.T) {
	t.Parallel()
	g := NewGraph[string]()
	if g.VertexCount() != 0 {
		t.Errorf("VertexCount = %d; want 0", g.VertexCount())
	}
	if g.EdgeCount() != 0 {
		t.Errorf("EdgeCount = %d; want 0", g.EdgeCount())
	}
}

func TestGraphAddVertex(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	g.AddVertex(2)
	g.AddVertex(1) // duplicate
	if g.VertexCount() != 2 {
		t.Errorf("VertexCount = %d; want 2", g.VertexCount())
	}
}

func TestGraphZeroValue(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	g.AddVertex(1)
	if !g.HasVertex(1) {
		t.Error("zero value AddVertex failed")
	}
}

func TestGraphAddEdge(t *testing.T) {
	t.Parallel()
	g := NewGraph[string]()
	g.AddEdge("a", "b")
	if !g.HasEdge("a", "b") {
		t.Error("HasEdge(a,b) = false")
	}
	if g.HasEdge("b", "a") {
		t.Error("HasEdge(b,a) = true; should be directed")
	}
	if g.EdgeCount() != 1 {
		t.Errorf("EdgeCount = %d; want 1", g.EdgeCount())
	}
}

func TestGraphHasVertexNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	if g.HasVertex(1) {
		t.Error("nil HasVertex = true")
	}
}

func TestGraphHasEdgeNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	if g.HasEdge(1, 2) {
		t.Error("nil HasEdge = true")
	}
}

func TestGraphHasEdgeMissingVertex(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	if g.HasEdge(1, 2) {
		t.Error("HasEdge(1,2) = true; 2 not in neighbours")
	}
	if g.HasEdge(3, 1) {
		t.Error("HasEdge(3,1) = true; 3 not in graph")
	}
}

func TestGraphNeighbors(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	n := g.Neighbors(1)
	sort.Ints(n)
	if len(n) != 2 || n[0] != 2 || n[1] != 3 {
		t.Errorf("Neighbors(1) = %v; want [2 3]", n)
	}
}

func TestGraphNeighborsNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	if n := g.Neighbors(1); n != nil {
		t.Errorf("nil Neighbors = %v; want nil", n)
	}
}

func TestGraphNeighborsMissing(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	if n := g.Neighbors(99); n != nil {
		t.Errorf("missing Neighbors = %v; want nil", n)
	}
}

func TestGraphVertices(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(3)
	g.AddVertex(1)
	g.AddVertex(2)
	v := g.Vertices()
	sort.Ints(v)
	if len(v) != 3 || v[0] != 1 || v[1] != 2 || v[2] != 3 {
		t.Errorf("Vertices = %v; want [1 2 3]", v)
	}
}

func TestGraphVerticesNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	if v := g.Vertices(); v != nil {
		t.Errorf("nil Vertices = %v; want nil", v)
	}
}

func TestGraphRemoveEdge(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.RemoveEdge(1, 2)
	if g.HasEdge(1, 2) {
		t.Error("HasEdge after RemoveEdge = true")
	}
}

func TestGraphRemoveEdgeNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	g.RemoveEdge(1, 2) // should not panic
}

func TestGraphRemoveEdgeMissing(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	g.RemoveEdge(99, 1) // missing source
}

func TestGraphRemoveVertex(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(3, 1)
	g.RemoveVertex(2)
	if g.HasVertex(2) {
		t.Error("HasVertex(2) after removal = true")
	}
	if g.HasEdge(1, 2) {
		t.Error("HasEdge(1,2) after removal = true")
	}
}

func TestGraphRemoveVertexNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	g.RemoveVertex(1) // should not panic
}

func TestGraphBFS(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(1, 3)
	g.AddEdge(2, 4)
	g.AddEdge(3, 4)
	var visited []int
	g.BFS(1, func(v int) bool {
		visited = append(visited, v)
		return true
	})
	if visited[0] != 1 {
		t.Errorf("BFS first = %d; want 1", visited[0])
	}
	if len(visited) != 4 {
		t.Errorf("BFS visited %d; want 4", len(visited))
	}
}

func TestGraphBFSEarlyStop(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	count := 0
	g.BFS(1, func(_ int) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Errorf("BFS stopped at %d; want 2", count)
	}
}

func TestGraphBFSNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	g.BFS(1, func(_ int) bool { return true }) // should not panic
}

func TestGraphBFSMissingStart(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	called := false
	g.BFS(99, func(_ int) bool { called = true; return true })
	if called {
		t.Error("BFS called fn for missing start")
	}
}

func TestGraphDFS(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	var visited []int
	g.DFS(1, func(v int) bool {
		visited = append(visited, v)
		return true
	})
	if visited[0] != 1 {
		t.Errorf("DFS first = %d; want 1", visited[0])
	}
	if len(visited) != 3 {
		t.Errorf("DFS visited %d; want 3", len(visited))
	}
}

func TestGraphDFSEarlyStop(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddEdge(1, 2)
	g.AddEdge(2, 3)
	g.AddEdge(3, 4)
	count := 0
	g.DFS(1, func(_ int) bool {
		count++
		return count < 2
	})
	if count != 2 {
		t.Errorf("DFS stopped at %d; want 2", count)
	}
}

func TestGraphDFSNil(t *testing.T) {
	t.Parallel()
	var g Graph[int]
	g.DFS(1, func(_ int) bool { return true }) // should not panic
}

func TestGraphDFSMissingStart(t *testing.T) {
	t.Parallel()
	g := NewGraph[int]()
	g.AddVertex(1)
	called := false
	g.DFS(99, func(_ int) bool { called = true; return true })
	if called {
		t.Error("DFS called fn for missing start")
	}
}

func BenchmarkGraphAddEdge(b *testing.B) {
	g := NewGraph[int]()
	for i := range b.N {
		g.AddEdge(i, i+1)
	}
}

func BenchmarkGraphBFS(b *testing.B) {
	g := NewGraph[int]()
	for i := range 1000 {
		g.AddEdge(i, i+1)
	}
	b.ResetTimer()
	for range b.N {
		g.BFS(0, func(_ int) bool { return true })
	}
}

func FuzzGraphAddEdge(f *testing.F) {
	f.Add(1, 2)
	f.Add(0, 0)
	f.Add(-1, 100)
	f.Fuzz(func(t *testing.T, u, v int) {
		g := NewGraph[int]()
		g.AddEdge(u, v)
		if !g.HasEdge(u, v) {
			t.Errorf("HasEdge(%d, %d) = false after AddEdge", u, v)
		}
	})
}
