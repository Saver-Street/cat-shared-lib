package types

import "testing"

func TestMatrixNew(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](3, 4)
	if m.Rows() != 3 {
		t.Errorf("Rows = %d; want 3", m.Rows())
	}
	if m.Cols() != 4 {
		t.Errorf("Cols = %d; want 4", m.Cols())
	}
}

func TestMatrixNewPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	NewMatrix[int](0, 5)
}

func TestMatrixGetSet(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 3)
	m.Set(1, 2, 42)
	if v := m.Get(1, 2); v != 42 {
		t.Errorf("Get(1,2) = %d; want 42", v)
	}
}

func TestMatrixGetPanic(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-bounds Get")
		}
	}()
	m.Get(5, 0)
}

func TestMatrixSetPanic(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for out-of-bounds Set")
		}
	}()
	m.Set(0, 5, 1)
}

func TestMatrixRow(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 3)
	m.Set(1, 0, 1)
	m.Set(1, 1, 2)
	m.Set(1, 2, 3)
	row := m.Row(1)
	if len(row) != 3 || row[0] != 1 || row[1] != 2 || row[2] != 3 {
		t.Errorf("Row(1) = %v; want [1 2 3]", row)
	}
	// Mutating should not affect matrix.
	row[0] = 99
	if m.Get(1, 0) != 1 {
		t.Error("Row mutation affected matrix")
	}
}

func TestMatrixRowPanic(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	m.Row(5)
}

func TestMatrixCol(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](3, 2)
	m.Set(0, 1, 10)
	m.Set(1, 1, 20)
	m.Set(2, 1, 30)
	col := m.Col(1)
	if len(col) != 3 || col[0] != 10 || col[1] != 20 || col[2] != 30 {
		t.Errorf("Col(1) = %v; want [10 20 30]", col)
	}
}

func TestMatrixColPanic(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	m.Col(5)
}

func TestMatrixFill(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	m.Fill(7)
	for r := range 2 {
		for c := range 2 {
			if v := m.Get(r, c); v != 7 {
				t.Errorf("Get(%d,%d) = %d after Fill(7)", r, c, v)
			}
		}
	}
}

func TestMatrixEach(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 3)
	m.Fill(1)
	sum := 0
	m.Each(func(_, _ int, v int) bool {
		sum += v
		return true
	})
	if sum != 6 {
		t.Errorf("Each sum = %d; want 6", sum)
	}
}

func TestMatrixEachEarlyStop(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](3, 3)
	m.Fill(1)
	count := 0
	m.Each(func(_, _ int, _ int) bool {
		count++
		return count < 4
	})
	if count != 4 {
		t.Errorf("Each stopped at %d; want 4", count)
	}
}

func TestMatrixFlat(t *testing.T) {
	t.Parallel()
	m := NewMatrix[int](2, 2)
	m.Set(0, 0, 1)
	m.Set(0, 1, 2)
	m.Set(1, 0, 3)
	m.Set(1, 1, 4)
	flat := m.Flat()
	if len(flat) != 4 || flat[0] != 1 || flat[1] != 2 || flat[2] != 3 || flat[3] != 4 {
		t.Errorf("Flat = %v; want [1 2 3 4]", flat)
	}
	// Mutating should not affect matrix.
	flat[0] = 99
	if m.Get(0, 0) != 1 {
		t.Error("Flat mutation affected matrix")
	}
}

func BenchmarkMatrixGetSet(b *testing.B) {
	m := NewMatrix[int](100, 100)
	for i := range b.N {
		r, c := i%100, (i/100)%100
		m.Set(r, c, i)
		_ = m.Get(r, c)
	}
}

func BenchmarkMatrixFill(b *testing.B) {
	m := NewMatrix[int](100, 100)
	for range b.N {
		m.Fill(0)
	}
}

func FuzzMatrixGetSet(f *testing.F) {
	f.Add(0, 0, 42)
	f.Add(1, 2, -1)
	f.Fuzz(func(t *testing.T, r, c, v int) {
		if r < 0 || r >= 10 || c < 0 || c >= 10 {
			return
		}
		m := NewMatrix[int](10, 10)
		m.Set(r, c, v)
		if got := m.Get(r, c); got != v {
			t.Errorf("Set/Get(%d, %d, %d) = %d", r, c, v, got)
		}
	})
}
