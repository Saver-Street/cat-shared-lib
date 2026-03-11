package scan

import (
	"errors"
	"testing"
)

// fuzzRows is a RowScanner that returns fuzzed string data.
type fuzzRows struct {
	data []string
	idx  int
}

func (r *fuzzRows) Next() bool {
	if r.idx < len(r.data) {
		r.idx++
		return true
	}
	return false
}

func (r *fuzzRows) Scan(dest ...any) error {
	val := r.data[r.idx-1]
	if len(dest) > 0 {
		*dest[0].(*string) = val
	}
	return nil
}

func (r *fuzzRows) Close()     {}
func (r *fuzzRows) Err() error { return nil }

type singleField struct{ Val string }

func scanSingleField(s *singleField) []any { return []any{&s.Val} }

func FuzzRows(f *testing.F) {
	f.Add("hello", 1)
	f.Add("", 0)
	f.Add("a,b,c", 3)
	f.Add("special\x00chars\ttab\nnewline", 1)

	f.Fuzz(func(t *testing.T, val string, n int) {
		if n < 0 || n > 100 {
			t.Skip()
		}
		data := make([]string, n)
		for i := range data {
			data[i] = val
		}
		rows := &fuzzRows{data: data}
		got, err := Rows[singleField](rows, scanSingleField)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != n {
			t.Errorf("got %d rows, want %d", len(got), n)
		}
	})
}

func FuzzFirst(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("\x00\xff")
	f.Add("a long string with unicode: 日本語")

	f.Fuzz(func(t *testing.T, val string) {
		rows := &fuzzRows{data: []string{val}}
		got, err := First[singleField](rows, scanSingleField)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Val != val {
			t.Errorf("got %q, want %q", got.Val, val)
		}
	})
}

func FuzzFirstEmpty(f *testing.F) {
	f.Add(0)
	f.Add(1)

	f.Fuzz(func(t *testing.T, n int) {
		if n != 0 {
			t.Skip()
		}
		rows := &fuzzRows{data: []string{}}
		_, err := First[singleField](rows, scanSingleField)
		if !errors.Is(err, ErrNoRows) {
			t.Errorf("expected ErrNoRows, got %v", err)
		}
	})
}

func FuzzRowsLimit(f *testing.F) {
	f.Add("val", 5, 3)
	f.Add("", 10, 0)
	f.Add("x", 0, -1)
	f.Add("data", 100, 50)

	f.Fuzz(func(t *testing.T, val string, n, limit int) {
		if n < 0 || n > 100 {
			t.Skip()
		}
		data := make([]string, n)
		for i := range data {
			data[i] = val
		}
		rows := &fuzzRows{data: data}
		got, err := RowsLimit[singleField](rows, scanSingleField, limit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if limit > 0 && len(got) > limit {
			t.Errorf("got %d rows, exceeds limit %d", len(got), limit)
		}
	})
}
