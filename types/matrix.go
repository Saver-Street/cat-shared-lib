package types

import "fmt"

// Matrix is a generic 2D matrix backed by a flat slice.
type Matrix[T any] struct {
	data []T
	rows int
	cols int
}

// NewMatrix creates an r×c matrix with zero-valued elements.
// Panics if rows or cols is less than 1.
func NewMatrix[T any](rows, cols int) *Matrix[T] {
	if rows < 1 || cols < 1 {
		panic("matrix: dimensions must be >= 1")
	}
	return &Matrix[T]{
		data: make([]T, rows*cols),
		rows: rows,
		cols: cols,
	}
}

// Get returns the element at row r, column c.
// Panics if out of bounds.
func (m *Matrix[T]) Get(r, c int) T {
	m.checkBounds(r, c)
	return m.data[r*m.cols+c]
}

// Set sets the element at row r, column c.
// Panics if out of bounds.
func (m *Matrix[T]) Set(r, c int, v T) {
	m.checkBounds(r, c)
	m.data[r*m.cols+c] = v
}

// Rows returns the number of rows.
func (m *Matrix[T]) Rows() int { return m.rows }

// Cols returns the number of columns.
func (m *Matrix[T]) Cols() int { return m.cols }

// Row returns a copy of row r.
func (m *Matrix[T]) Row(r int) []T {
	if r < 0 || r >= m.rows {
		panic(fmt.Sprintf("matrix: row %d out of range [0, %d)", r, m.rows))
	}
	row := make([]T, m.cols)
	copy(row, m.data[r*m.cols:(r+1)*m.cols])
	return row
}

// Col returns a copy of column c.
func (m *Matrix[T]) Col(c int) []T {
	if c < 0 || c >= m.cols {
		panic(fmt.Sprintf("matrix: col %d out of range [0, %d)", c, m.cols))
	}
	col := make([]T, m.rows)
	for r := range m.rows {
		col[r] = m.data[r*m.cols+c]
	}
	return col
}

// Fill sets every element to v.
func (m *Matrix[T]) Fill(v T) {
	for i := range m.data {
		m.data[i] = v
	}
}

// Each calls fn for every element with its row and column index.
// If fn returns false, iteration stops.
func (m *Matrix[T]) Each(fn func(r, c int, v T) bool) {
	for r := range m.rows {
		for c := range m.cols {
			if !fn(r, c, m.data[r*m.cols+c]) {
				return
			}
		}
	}
}

// Flat returns a copy of the underlying flat representation (row-major).
func (m *Matrix[T]) Flat() []T {
	cp := make([]T, len(m.data))
	copy(cp, m.data)
	return cp
}

func (m *Matrix[T]) checkBounds(r, c int) {
	if r < 0 || r >= m.rows || c < 0 || c >= m.cols {
		panic(fmt.Sprintf("matrix: index (%d, %d) out of range [%d×%d]", r, c, m.rows, m.cols))
	}
}
