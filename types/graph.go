package types

// Graph is a directed graph with comparable vertex keys.
// The zero value is ready to use.
type Graph[T comparable] struct {
	adj map[T]map[T]struct{}
}

// NewGraph returns an initialised Graph.
func NewGraph[T comparable]() *Graph[T] {
	return &Graph[T]{adj: make(map[T]map[T]struct{})}
}

// AddVertex adds a vertex to the graph. If it already exists this is a no-op.
func (g *Graph[T]) AddVertex(v T) {
	if g.adj == nil {
		g.adj = make(map[T]map[T]struct{})
	}
	if _, ok := g.adj[v]; !ok {
		g.adj[v] = make(map[T]struct{})
	}
}

// AddEdge adds a directed edge from u to v.
// Both vertices are created if they don't exist.
func (g *Graph[T]) AddEdge(u, v T) {
	g.AddVertex(u)
	g.AddVertex(v)
	g.adj[u][v] = struct{}{}
}

// HasVertex reports whether v is in the graph.
func (g *Graph[T]) HasVertex(v T) bool {
	if g.adj == nil {
		return false
	}
	_, ok := g.adj[v]
	return ok
}

// HasEdge reports whether a directed edge from u to v exists.
func (g *Graph[T]) HasEdge(u, v T) bool {
	if g.adj == nil {
		return false
	}
	neighbours, ok := g.adj[u]
	if !ok {
		return false
	}
	_, ok = neighbours[v]
	return ok
}

// Neighbors returns the direct successors of v.
func (g *Graph[T]) Neighbors(v T) []T {
	if g.adj == nil {
		return nil
	}
	neighbours, ok := g.adj[v]
	if !ok {
		return nil
	}
	result := make([]T, 0, len(neighbours))
	for n := range neighbours {
		result = append(result, n)
	}
	return result
}

// Vertices returns all vertices in the graph.
func (g *Graph[T]) Vertices() []T {
	if g.adj == nil {
		return nil
	}
	verts := make([]T, 0, len(g.adj))
	for v := range g.adj {
		verts = append(verts, v)
	}
	return verts
}

// VertexCount returns the number of vertices.
func (g *Graph[T]) VertexCount() int {
	return len(g.adj)
}

// EdgeCount returns the number of directed edges.
func (g *Graph[T]) EdgeCount() int {
	n := 0
	for _, neighbours := range g.adj {
		n += len(neighbours)
	}
	return n
}

// RemoveEdge removes the directed edge from u to v.
func (g *Graph[T]) RemoveEdge(u, v T) {
	if g.adj == nil {
		return
	}
	if neighbours, ok := g.adj[u]; ok {
		delete(neighbours, v)
	}
}

// RemoveVertex removes v and all edges to/from it.
func (g *Graph[T]) RemoveVertex(v T) {
	if g.adj == nil {
		return
	}
	delete(g.adj, v)
	for _, neighbours := range g.adj {
		delete(neighbours, v)
	}
}

// BFS performs a breadth-first search from start, calling fn for each
// visited vertex. If fn returns false, traversal stops.
func (g *Graph[T]) BFS(start T, fn func(T) bool) {
	if g.adj == nil {
		return
	}
	if _, ok := g.adj[start]; !ok {
		return
	}
	visited := make(map[T]struct{})
	queue := []T{start}
	visited[start] = struct{}{}
	for len(queue) > 0 {
		v := queue[0]
		queue = queue[1:]
		if !fn(v) {
			return
		}
		for n := range g.adj[v] {
			if _, seen := visited[n]; !seen {
				visited[n] = struct{}{}
				queue = append(queue, n)
			}
		}
	}
}

// DFS performs a depth-first search from start, calling fn for each
// visited vertex. If fn returns false, traversal stops.
func (g *Graph[T]) DFS(start T, fn func(T) bool) {
	if g.adj == nil {
		return
	}
	if _, ok := g.adj[start]; !ok {
		return
	}
	visited := make(map[T]struct{})
	g.dfs(start, visited, fn)
}

func (g *Graph[T]) dfs(v T, visited map[T]struct{}, fn func(T) bool) bool {
	visited[v] = struct{}{}
	if !fn(v) {
		return false
	}
	for n := range g.adj[v] {
		if _, seen := visited[n]; !seen {
			if !g.dfs(n, visited, fn) {
				return false
			}
		}
	}
	return true
}
