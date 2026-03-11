package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStreamWriterBasic(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	if err := sw.Write(map[string]int{"a": 1}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sw.Write(map[string]int{"b": 2}); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	body := rec.Body.String()
	// json.Encoder adds newlines after each element
	body = strings.ReplaceAll(body, "\n", "")
	want := `[{"a":1},{"b":2}]`
	if body != want {
		t.Errorf("body = %q; want %q", body, want)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q; want application/json", ct)
	}
	if sw.Count() != 2 {
		t.Errorf("Count() = %d; want 2", sw.Count())
	}
}

func TestStreamWriterEmpty(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)
	if err := sw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	body := rec.Body.String()
	if body != "[]" {
		t.Errorf("body = %q; want []", body)
	}
	if sw.Count() != 0 {
		t.Errorf("Count() = %d; want 0", sw.Count())
	}
}

func TestStreamWriterSingle(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	if err := sw.Write("hello"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	body := strings.ReplaceAll(rec.Body.String(), "\n", "")
	if body != `["hello"]` {
		t.Errorf("body = %q; want [\"hello\"]", body)
	}
}

type failWriter struct {
	header http.Header
	failed bool
}

func (f *failWriter) Header() http.Header { return f.header }
func (f *failWriter) Write([]byte) (int, error) {
	if f.failed {
		return 0, http.ErrAbortHandler
	}
	f.failed = true
	return 0, http.ErrAbortHandler
}
func (f *failWriter) WriteHeader(int) {}

func TestStreamWriterWriteError(t *testing.T) {
	t.Parallel()
	fw := &failWriter{header: http.Header{}}
	sw := NewStreamWriter(fw)

	// First write should fail because NewStreamWriter already failed on "["
	err := sw.Write("data")
	if err == nil {
		t.Error("expected error on Write after failed open bracket")
	}
	// Close should also propagate error
	if sw.Close() == nil {
		t.Error("expected error on Close")
	}
}

type nonFlusherWriter struct {
	header http.Header
	body   []byte
}

func (n *nonFlusherWriter) Header() http.Header { return n.header }
func (n *nonFlusherWriter) Write(b []byte) (int, error) {
	n.body = append(n.body, b...)
	return len(b), nil
}
func (n *nonFlusherWriter) WriteHeader(int) {}

func TestStreamWriterNoFlusher(t *testing.T) {
	t.Parallel()
	nf := &nonFlusherWriter{header: http.Header{}}
	sw := NewStreamWriter(nf)

	if err := sw.Write(42); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := sw.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	body := strings.ReplaceAll(string(nf.body), "\n", "")
	if body != "[42]" {
		t.Errorf("body = %q; want [42]", body)
	}
}

type commaFailWriter struct {
	header http.Header
	count  int
}

func (c *commaFailWriter) Header() http.Header { return c.header }
func (c *commaFailWriter) Write(b []byte) (int, error) {
	c.count++
	if c.count == 3 {
		// fail on the comma write (3rd call: "[", first encode, then ",")
		return 0, http.ErrAbortHandler
	}
	return len(b), nil
}
func (c *commaFailWriter) WriteHeader(int) {}

func TestStreamWriterCommaError(t *testing.T) {
	t.Parallel()
	cw := &commaFailWriter{header: http.Header{}}
	sw := NewStreamWriter(cw)

	// First write succeeds
	if err := sw.Write(1); err != nil {
		t.Fatalf("first Write: %v", err)
	}
	// Second write should fail on comma
	if err := sw.Write(2); err == nil {
		t.Error("expected error on comma write")
	}
}

// Test that a channel-based encoding error (like func values) propagates
func TestStreamWriterEncodeError(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	sw := NewStreamWriter(rec)

	// Functions can't be JSON-encoded
	err := sw.Write(func() {})
	if err == nil {
		t.Error("expected JSON encode error")
	}
}

func BenchmarkStreamWriter(b *testing.B) {
	type item struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	for range b.N {
		rec := httptest.NewRecorder()
		sw := NewStreamWriter(rec)
		for i := range 10 {
			_ = sw.Write(item{ID: i, Name: "test"})
		}
		_ = sw.Close()
	}
}

func FuzzStreamWriter(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("special \"chars\" & <tags>")
	f.Fuzz(func(t *testing.T, s string) {
		rec := httptest.NewRecorder()
		sw := NewStreamWriter(rec)
		_ = sw.Write(s)
		_ = sw.Close()
	})
}
