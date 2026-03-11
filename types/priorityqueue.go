package types

import "container/heap"

// PriorityQueue is a generic min-heap priority queue. Items with lower
// priority values are dequeued first. Use a negated priority for max-heap
// behavior.
type PriorityQueue[T any] struct {
	h *pqHeap[T]
}

// NewPriorityQueue returns an empty priority queue.
func NewPriorityQueue[T any]() *PriorityQueue[T] {
	h := &pqHeap[T]{}
	heap.Init(h)
	return &PriorityQueue[T]{h: h}
}

// Push adds an item with the given priority.
func (pq *PriorityQueue[T]) Push(value T, priority int) {
	heap.Push(pq.h, pqItem[T]{value: value, priority: priority})
}

// Pop removes and returns the item with the lowest priority value.
// It panics if the queue is empty.
func (pq *PriorityQueue[T]) Pop() T {
	item := heap.Pop(pq.h).(pqItem[T])
	return item.value
}

// Peek returns the item with the lowest priority value without removing it.
// It panics if the queue is empty.
func (pq *PriorityQueue[T]) Peek() T {
	return (*pq.h)[0].value
}

// Len returns the number of items in the queue.
func (pq *PriorityQueue[T]) Len() int {
	return pq.h.Len()
}

// IsEmpty reports whether the queue has no items.
func (pq *PriorityQueue[T]) IsEmpty() bool {
	return pq.h.Len() == 0
}

// pqItem is an internal wrapper holding a value and its priority.
type pqItem[T any] struct {
	value    T
	priority int
}

// pqHeap implements heap.Interface for pqItem elements.
type pqHeap[T any] []pqItem[T]

func (h pqHeap[T]) Len() int           { return len(h) }
func (h pqHeap[T]) Less(i, j int) bool { return h[i].priority < h[j].priority }
func (h pqHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *pqHeap[T]) Push(x any)        { *h = append(*h, x.(pqItem[T])) }
func (h *pqHeap[T]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}
