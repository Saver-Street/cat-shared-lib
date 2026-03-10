package scan

import (
	"errors"
	"testing"
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
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].Name != "a" || items[1].Name != "b" {
		t.Errorf("got %v, want a,b", items)
	}
	if !rows.closed {
		t.Error("rows not closed")
	}
}

func TestRows_Empty(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	items, err := Rows[item](rows, scanItem)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Errorf("got %d items, want 0", len(items))
	}
}

func TestRows_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := Rows[item](rows, scanItem)
	if err == nil || err.Error() != "scan fail" {
		t.Errorf("got err %v, want scan fail", err)
	}
}

func TestRows_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := Rows[item](rows, scanItem)
	if err == nil || err.Error() != "iter fail" {
		t.Errorf("got err %v, want iter fail", err)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "hello" || result.Val != "world" {
		t.Errorf("got %+v, want {hello world}", result)
	}
}

func TestRow_Error(t *testing.T) {
	row := &mockRow{err: errors.New("no rows")}
	_, err := Row[item](row, scanItem)
	if err == nil || err.Error() != "no rows" {
		t.Errorf("got err %v, want no rows", err)
	}
}

func TestRows_LargeDataSet(t *testing.T) {
	data := make([][]any, 100)
	for i := range data {
		data[i] = []any{"name", "val"}
	}
	rows := &mockRows{data: data}
	items, err := Rows[item](rows, scanItem)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 100 {
		t.Fatalf("got %d items, want 100", len(items))
	}
}

func TestRows_SingleRow(t *testing.T) {
	rows := &mockRows{data: [][]any{{"only", "one"}}}
	items, err := Rows[item](rows, scanItem)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Name != "only" {
		t.Errorf("got %q, want only", items[0].Name)
	}
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
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != 1 || items[1].ID != 2 {
		t.Errorf("got IDs %d,%d, want 1,2", items[0].ID, items[1].ID)
	}
}

func TestRow_DifferentGenericType(t *testing.T) {
	row := &mockRow{data: []any{"hello", "world"}}
	// Reuse item type but verify zero-value on error
	_, err := Row[item](&mockRow{err: errors.New("fail")}, scanItem)
	if err == nil {
		t.Fatal("expected error")
	}
	// Success with mockRow
	result, err := Row[item](row, scanItem)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "hello" {
		t.Errorf("got %q, want hello", result.Name)
	}
}

func TestRows_CloseCalledOnScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	Rows[item](rows, scanItem)
	if !rows.closed {
		t.Error("rows should be closed even after scan error")
	}
}

func TestRows_NilRows(t *testing.T) {
	_, err := Rows[item](nil, scanItem)
	if err == nil || err.Error() != "scan.Rows: rows must not be nil" {
		t.Errorf("got err %v, want nil rows error", err)
	}
}

func TestRows_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	_, err := Rows[item](rows, nil)
	if err == nil || err.Error() != "scan.Rows: scanFn must not be nil" {
		t.Errorf("got err %v, want nil scanFn error", err)
	}
}

func TestRow_NilRow(t *testing.T) {
	_, err := Row[item](nil, scanItem)
	if err == nil || err.Error() != "scan.Row: row must not be nil" {
		t.Errorf("got err %v, want nil row error", err)
	}
}

func TestRow_NilScanFn(t *testing.T) {
	row := &mockRow{data: []any{"hello", "world"}}
	_, err := Row[item](row, nil)
	if err == nil || err.Error() != "scan.Row: scanFn must not be nil" {
		t.Errorf("got err %v, want nil scanFn error", err)
	}
}

func TestFirst_Success(t *testing.T) {
	rows := &mockRows{data: [][]any{{"first", "val"}, {"second", "val2"}}}
	result, err := First[item](rows, scanItem)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "first" {
		t.Errorf("got %q, want first", result.Name)
	}
	if !rows.closed {
		t.Error("rows should be closed after First")
	}
}

func TestFirst_Empty(t *testing.T) {
	rows := &mockRows{data: [][]any{}}
	_, err := First[item](rows, scanItem)
	if err == nil {
		t.Fatal("expected error for empty rows")
	}
	if err.Error() != "scan.First: no rows in result set" {
		t.Errorf("got %q, want 'no rows in result set'", err.Error())
	}
}

func TestFirst_NilRows(t *testing.T) {
	_, err := First[item](nil, scanItem)
	if err == nil || err.Error() != "scan.First: rows must not be nil" {
		t.Errorf("got err %v, want nil rows error", err)
	}
}

func TestFirst_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "b"}}}
	_, err := First[item](rows, nil)
	if err == nil || err.Error() != "scan.First: scanFn must not be nil" {
		t.Errorf("got err %v, want nil scanFn error", err)
	}
}

func TestFirst_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := First[item](rows, scanItem)
	if err == nil || err.Error() != "scan fail" {
		t.Errorf("got err %v, want scan fail", err)
	}
}

func TestFirst_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := First[item](rows, scanItem)
	if err == nil || err.Error() != "iter fail" {
		t.Errorf("got err %v, want iter fail", err)
	}
}

func TestRowsLimit_Basic(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}, {"b", "2"}, {"c", "3"}}}
	got, err := RowsLimit[item](rows, scanItem, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("len = %d, want 2", len(got))
	}
}

func TestRowsLimit_ZeroLimit_ReturnsAll(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}, {"b", "2"}, {"c", "3"}}}
	got, err := RowsLimit[item](rows, scanItem, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Errorf("len = %d, want 3", len(got))
	}
}

func TestRowsLimit_NegativeLimit_ReturnsAll(t *testing.T) {
	rows := &mockRows{data: [][]any{{"x", "9"}}}
	got, err := RowsLimit[item](rows, scanItem, -1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Errorf("len = %d, want 1", len(got))
	}
}

func TestRowsLimit_NilRows(t *testing.T) {
	_, err := RowsLimit[item](nil, scanItem, 5)
	if err == nil {
		t.Fatal("expected error for nil rows")
	}
}

func TestRowsLimit_NilScanFn(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}}
	_, err := RowsLimit[item](rows, nil, 5)
	if err == nil {
		t.Fatal("expected error for nil scanFn")
	}
}

func TestRowsLimit_IterError(t *testing.T) {
	rows := &mockRows{data: [][]any{}, iterErr: errors.New("iter fail")}
	_, err := RowsLimit[item](rows, scanItem, 5)
	if err == nil || err.Error() != "iter fail" {
		t.Errorf("got err %v, want 'iter fail'", err)
	}
}

func TestRowsLimit_ScanError(t *testing.T) {
	rows := &mockRows{data: [][]any{{"a", "1"}}, scanErr: errors.New("scan fail")}
	_, err := RowsLimit[item](rows, scanItem, 5)
	if err == nil {
		t.Fatal("expected scan error from RowsLimit")
	}
}
