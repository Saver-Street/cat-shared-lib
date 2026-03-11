package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONLWriterSingle(t *testing.T) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)

	if err := jw.Write(map[string]int{"id": 1}); err != nil {
		t.Fatal(err)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/x-ndjson" {
		t.Fatalf("Content-Type = %q, want application/x-ndjson", ct)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	got := strings.TrimSpace(w.Body.String())
	if got != `{"id":1}` {
		t.Fatalf("body = %q, want {\"id\":1}", got)
	}
}

func TestJSONLWriterMultiple(t *testing.T) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)

	_ = jw.Write(map[string]int{"id": 1})
	_ = jw.Write(map[string]int{"id": 2})
	_ = jw.Write(map[string]int{"id": 3})

	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
}

func TestJSONLWriterMarshalError(t *testing.T) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)

	err := jw.Write(func() {})
	if err == nil {
		t.Fatal("expected error for unmarshalable value")
	}
}

func TestWriteSlice(t *testing.T) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)

	items := []string{"a", "b", "c"}
	if err := WriteSlice(jw, items); err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(w.Body.String()), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}
	if lines[0] != `"a"` {
		t.Fatalf("line[0] = %q, want \"a\"", lines[0])
	}
}

func TestWriteSliceEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)

	if err := WriteSlice(jw, []int{}); err != nil {
		t.Fatal(err)
	}
	if w.Body.Len() != 0 {
		t.Fatalf("body should be empty for empty slice")
	}
}

type flusherRecorder struct {
	*httptest.ResponseRecorder
	flushed int
}

func (f *flusherRecorder) Flush() {
	f.flushed++
}

func TestJSONLWriterFlushes(t *testing.T) {
	rec := httptest.NewRecorder()
	fr := &flusherRecorder{ResponseRecorder: rec}
	jw := NewJSONLWriter(fr)

	_ = jw.Write("test")
	if fr.flushed != 1 {
		t.Fatalf("flushed = %d, want 1", fr.flushed)
	}
	_ = jw.Write("test2")
	if fr.flushed != 2 {
		t.Fatalf("flushed = %d, want 2", fr.flushed)
	}
}

func BenchmarkJSONLWrite(b *testing.B) {
	w := httptest.NewRecorder()
	jw := NewJSONLWriter(w)
	item := map[string]int{"id": 1, "value": 42}
	for b.Loop() {
		_ = jw.Write(item)
	}
}

func BenchmarkWriteSlice(b *testing.B) {
	items := make([]int, 100)
	for i := range items {
		items[i] = i
	}
	for b.Loop() {
		w := httptest.NewRecorder()
		jw := NewJSONLWriter(w)
		_ = WriteSlice(jw, items)
	}
}

func FuzzJSONLWrite(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("test data")

	f.Fuzz(func(t *testing.T, s string) {
		w := httptest.NewRecorder()
		jw := NewJSONLWriter(w)
		_ = jw.Write(s)
	})
}

type errorWriter struct {
	http.ResponseWriter
}

func (e *errorWriter) Header() http.Header {
	return http.Header{}
}

func (e *errorWriter) WriteHeader(int) {}

func (e *errorWriter) Write([]byte) (int, error) {
	return 0, http.ErrHandlerTimeout
}

func TestJSONLWriteError(t *testing.T) {
	jw := NewJSONLWriter(&errorWriter{})
	err := jw.Write("test")
	if err == nil {
		t.Fatal("expected write error")
	}
}

func TestWriteSliceError(t *testing.T) {
	jw := NewJSONLWriter(&errorWriter{})
	err := WriteSlice(jw, []string{"a", "b"})
	if err == nil {
		t.Fatal("expected error from WriteSlice")
	}
}
