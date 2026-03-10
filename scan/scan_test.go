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
