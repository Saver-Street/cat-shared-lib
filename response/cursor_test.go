package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEncodeCursor(t *testing.T) {
	tests := []struct {
		name string
		val  any
		ok   bool
	}{
		{"string", "abc", true},
		{"int", 42, true},
		{"struct", struct{ ID int }{1}, true},
		{"nil", nil, true},
		{"func fails", func() {}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EncodeCursor(tt.val)
			if tt.ok && err != nil {
				t.Fatalf("EncodeCursor() error = %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("EncodeCursor() expected error")
			}
		})
	}
}

func TestDecodeCursor(t *testing.T) {
	t.Run("round trip string", func(t *testing.T) {
		cur, err := EncodeCursor("hello")
		if err != nil {
			t.Fatal(err)
		}
		var got string
		if err := DecodeCursor(cur, &got); err != nil {
			t.Fatal(err)
		}
		if got != "hello" {
			t.Fatalf("got %q, want hello", got)
		}
	})

	t.Run("round trip struct", func(t *testing.T) {
		type C struct {
			ID   int
			Name string
		}
		orig := C{ID: 42, Name: "test"}
		cur, err := EncodeCursor(orig)
		if err != nil {
			t.Fatal(err)
		}
		var got C
		if err := DecodeCursor(cur, &got); err != nil {
			t.Fatal(err)
		}
		if got != orig {
			t.Fatalf("got %+v, want %+v", got, orig)
		}
	})

	t.Run("invalid base64", func(t *testing.T) {
		var s string
		if err := DecodeCursor("not!valid!base64!", &s); err == nil {
			t.Fatal("expected error for invalid base64")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		// Valid base64 but not valid JSON for target type
		cur := "aGVsbG8=" // "hello" in base64
		var n int
		if err := DecodeCursor(cur, &n); err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestWriteCursorPage(t *testing.T) {
	w := httptest.NewRecorder()
	page := CursorPage[string]{
		Data:       []string{"a", "b"},
		NextCursor: "next123",
		HasMore:    true,
	}
	WriteCursorPage(w, page)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", ct)
	}

	var got CursorPage[string]
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Data) != 2 || got.Data[0] != "a" {
		t.Fatalf("data = %v, want [a b]", got.Data)
	}
	if got.NextCursor != "next123" {
		t.Fatalf("next_cursor = %q, want next123", got.NextCursor)
	}
	if !got.HasMore {
		t.Fatal("has_more = false, want true")
	}
}

func TestWriteCursorPageEmpty(t *testing.T) {
	w := httptest.NewRecorder()
	page := CursorPage[int]{
		Data:    []int{},
		HasMore: false,
	}
	WriteCursorPage(w, page)

	var got CursorPage[int]
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.HasMore {
		t.Fatal("has_more = true, want false")
	}
	if got.NextCursor != "" {
		t.Fatalf("next_cursor = %q, want empty", got.NextCursor)
	}
}

func TestWriteCursorPageWithPrev(t *testing.T) {
	w := httptest.NewRecorder()
	page := CursorPage[int]{
		Data:       []int{3, 4},
		PrevCursor: "prev456",
		NextCursor: "next789",
		HasMore:    true,
	}
	WriteCursorPage(w, page)

	var got CursorPage[int]
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.PrevCursor != "prev456" {
		t.Fatalf("prev_cursor = %q, want prev456", got.PrevCursor)
	}
}

func TestCursorPageJSON(t *testing.T) {
	page := CursorPage[string]{
		Data:       []string{"x"},
		NextCursor: "abc",
		HasMore:    true,
	}
	b, err := json.Marshal(page)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["prev_cursor"]; ok {
		t.Fatal("prev_cursor should be omitted when empty")
	}
}

func BenchmarkEncodeCursor(b *testing.B) {
	type C struct {
		ID   int
		Name string
	}
	v := C{ID: 1, Name: "test"}
	for b.Loop() {
		EncodeCursor(v)
	}
}

func BenchmarkDecodeCursor(b *testing.B) {
	cur, _ := EncodeCursor(42)
	var n int
	for b.Loop() {
		DecodeCursor(cur, &n)
	}
}

func FuzzEncodeDecode(f *testing.F) {
	f.Add("hello")
	f.Add("")
	f.Add("test-cursor-value")

	f.Fuzz(func(t *testing.T, s string) {
		cur, err := EncodeCursor(s)
		if err != nil {
			t.Fatal(err)
		}
		var got string
		if err := DecodeCursor(cur, &got); err != nil {
			t.Fatal(err)
		}
		if got != s {
			t.Fatalf("round trip failed: got %q, want %q", got, s)
		}
	})
}
