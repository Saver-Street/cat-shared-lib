package scan

import (
	"errors"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

// mockRows implements RowScanner for testing.
type mockRows struct {
	data    [][]any
	idx     int
	closed  bool
	scanErr error
	iterErr error
}

func (m *mockRows) Next() bool {
	if m.idx < len(m.data) {
		m.idx++
		return true
	}
	return false
}

func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	row := m.data[m.idx-1]
	for i, d := range dest {
		if i < len(row) {
			ptr := d.(*string)
			*ptr = row[i].(string)
		}
	}
	return nil
}

func (m *mockRows) Close()     { m.closed = true }
func (m *mockRows) Err() error { return m.iterErr }

type item struct {
	Name string
	Val  string
}

func scanItem(t *item) []any {
	return []any{&t.Name, &t.Val}
}

func TestRows_MultipleRows(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}, {"b", "2"}}}
	items, err := Rows[item](rows, scanItem)
	testkit.RequireNoError(t, err)
	testkit.RequireLen(t, items, 2)
	testkit.AssertEqual(t, items[0].Name, "a")
	testkit.AssertEqual(t, items[1].Name, "b")
	testkit.AssertTrue(t, rows.closed)
}

func TestRows_Empty(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	items, err := Rows[item](rows, scanItem)
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, items, 0)
}

func TestRows_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := Rows[item](rows, scanItem)
	testkit.AssertErrorContains(t, err, "scan fail")
}

func TestRows_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := Rows[item](rows, scanItem)
	testkit.AssertErrorContains(t, err, "iter fail")
}

// mockRow implements SingleRowScanner for testing.
type mockRow struct {
	data []any
	err  error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, d := range dest {
		if i < len(m.data) {
			ptr := d.(*string)
			*ptr = m.data[i].(string)
		}
	}
	return nil
}

func TestRow_Success(t *testing.T) {
	row := &mockRow{data: []any{"hello", "world"}}
	result, err := Row[item](row, scanItem)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, result.Name, "hello")
	testkit.AssertEqual(t, result.Val, "world")
}

func TestRow_Error(t *testing.T) {
	row := &mockRow{err: errors.New("no rows")}
	_, err := Row[item](row, scanItem)
	testkit.AssertErrorContains(t, err, "no rows")
}

func TestRows_LargeDataSet(t *testing.T) {
	data := make([][]any, 100)
	for i := range data {
		data[i] = []any{"name", "val"}
	}
	rows := &mockRows{data: data}
	items, err := Rows[item](rows, scanItem)
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, items, 100)
}

func TestRows_SingleRow(t *testing.T) {
	rows := &mockRows{data: [][]any{{"only", "one"}}}
	items, err := Rows[item](rows, scanItem)
	testkit.RequireNoError(t, err)
	testkit.RequireLen(t, items, 1)
	testkit.AssertEqual(t, items[0].Name, "only")
}

func BenchmarkRows(b *testing.B) {
	data := make([][]any, 10)
	for i := range data {
		data[i] = []any{"name", "val"}
	}
	for b.Loop() {
		rows := &mockRows{data: data}
		Rows[item](rows, scanItem)
	}
}

func BenchmarkRow(b *testing.B) {
	for b.Loop() {
		row := &mockRow{data: []any{"hello", "world"}}
		Row[item](row, scanItem)
	}
}

func BenchmarkFirst(b *testing.B) {
	data := [][]any{{"name", "val"}, {"name2", "val2"}}
	for b.Loop() {
		rows := &mockRows{data: data}
		First[item](rows, scanItem)
	}
}

func BenchmarkRowsLimit(b *testing.B) {
	data := make([][]any, 10)
	for i := range data {
		data[i] = []any{"name", "val"}
	}
	for b.Loop() {
		rows := &mockRows{data: data}
		RowsLimit[item](rows, scanItem, 5)
	}
}

// intItem verifies generics work with different types.
type intItem struct {
	ID   int
	Name string
}

// mockIntRows implements RowScanner for int-based items.
type mockIntRows struct {
	data [][]any
	idx  int
}

func (m *mockIntRows) Next() bool {
	if m.idx < len(m.data) {
		m.idx++
		return true
	}
	return false
}

func (m *mockIntRows) Scan(dest ...any) error {
	row := m.data[m.idx-1]
	*dest[0].(*int) = row[0].(int)
	*dest[1].(*string) = row[1].(string)
	return nil
}

func (m *mockIntRows) Close()     {}
func (m *mockIntRows) Err() error { return nil }

func TestRows_DifferentGenericType(t *testing.T) {
	rows := &mockIntRows{data: [][]any{{1, "first"}, {2, "second"}}}
	items, err := Rows[intItem](rows, func(it *intItem) []any {
		return []any{&it.ID, &it.Name}
	})
	testkit.RequireNoError(t, err)
	testkit.RequireLen(t, items, 2)
	testkit.AssertEqual(t, items[0].ID, 1)
	testkit.AssertEqual(t, items[1].ID, 2)
}

func TestRow_DifferentGenericType(t *testing.T) {
	row := &mockRow{data: []any{"hello", "world"}}
	// Reuse item type but verify zero-value on error
	_, err := Row[item](&mockRow{err: errors.New("fail")}, scanItem)
	testkit.AssertError(t, err)
	// Success with mockRow
	result, err := Row[item](row, scanItem)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, result.Name, "hello")
}

func TestRows_CloseCalledOnScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	Rows[item](rows, scanItem)
	testkit.AssertTrue(t, rows.closed)
}

func TestRows_NilRows(t *testing.T) {
	_, err := Rows[item](nil, scanItem)
	testkit.AssertErrorContains(t, err, "rows must not be nil")
}

func TestRows_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	_, err := Rows[item](rows, nil)
	testkit.AssertErrorContains(t, err, "scanFn must not be nil")
}

func TestRow_NilRow(t *testing.T) {
	_, err := Row[item](nil, scanItem)
	testkit.AssertErrorContains(t, err, "row must not be nil")
}

func TestRow_NilScanFn(t *testing.T) {
	row := &mockRow{data: []any{"hello", "world"}}
	_, err := Row[item](row, nil)
	testkit.AssertErrorContains(t, err, "scanFn must not be nil")
}

func TestFirst_Success(t *testing.T) {
	rows := &mockRows{data: [][]any{{"first", "val"}, {"second", "val2"}}}
	result, err := First[item](rows, scanItem)
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, result.Name, "first")
	testkit.AssertTrue(t, rows.closed)
}

func TestFirst_Empty(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	_, err := First[item](rows, scanItem)
	testkit.AssertError(t, err)
	testkit.AssertErrorIs(t, err, ErrNoRows)
}

func TestFirst_NilRows(t *testing.T) {
	_, err := First[item](nil, scanItem)
	testkit.AssertErrorContains(t, err, "rows must not be nil")
}

func TestFirst_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "b"}}}
	_, err := First[item](rows, nil)
	testkit.AssertErrorContains(t, err, "scanFn must not be nil")
}

func TestFirst_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := First[item](rows, scanItem)
	testkit.AssertErrorContains(t, err, "scan fail")
}

func TestFirst_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := First[item](rows, scanItem)
	testkit.AssertErrorContains(t, err, "iter fail")
}

func TestRowsLimit_Basic(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}, {"b", "2"}, {"c", "3"}}}
	got, err := RowsLimit[item](rows, scanItem, 2)
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, got, 2)
}

func TestRowsLimit_ZeroLimit_ReturnsAll(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}, {"b", "2"}, {"c", "3"}}}
	got, err := RowsLimit[item](rows, scanItem, 0)
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, got, 3)
}

func TestRowsLimit_NegativeLimit_ReturnsAll(t *testing.T) {
	rows := &mockRows{data: [][]any{{"x", "9"}}}
	got, err := RowsLimit[item](rows, scanItem, -1)
	testkit.RequireNoError(t, err)
	testkit.AssertLen(t, got, 1)
}

func TestRowsLimit_NilRows(t *testing.T) {
	_, err := RowsLimit[item](nil, scanItem, 5)
	testkit.AssertError(t, err)
}

func TestRowsLimit_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}}
	_, err := RowsLimit[item](rows, nil, 5)
	testkit.AssertError(t, err)
}

func TestRowsLimit_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := RowsLimit[item](rows, scanItem, 5)
	testkit.AssertErrorContains(t, err, "iter fail")
}

func TestRowsLimit_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := RowsLimit[item](rows, scanItem, 5)
	testkit.AssertError(t, err)
}

func TestValue_Success(t *testing.T) {
row := &mockRow{data: []any{"hello"}}
got, err := Value[string](row)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, got, "hello")
}

func TestValue_NilRow(t *testing.T) {
_, err := Value[string](nil)
testkit.AssertError(t, err)
testkit.AssertErrorContains(t, err, "row must not be nil")
}

func TestValue_ScanError(t *testing.T) {
row := &mockRow{err: errors.New("no rows")}
_, err := Value[string](row)
testkit.AssertError(t, err)
}
