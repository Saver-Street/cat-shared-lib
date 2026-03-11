package types

import (
	"encoding/json"
	"testing"
)

func TestLeft(t *testing.T) {
	t.Parallel()
	e := Left[string, int]("error")
	if !e.IsLeft() {
		t.Error("IsLeft() = false; want true")
	}
	if e.IsRight() {
		t.Error("IsRight() = true; want false")
	}
	l, ok := e.LeftVal()
	if !ok || l != "error" {
		t.Errorf("LeftVal() = (%q, %v); want (error, true)", l, ok)
	}
	_, ok = e.RightVal()
	if ok {
		t.Error("RightVal() ok = true; want false")
	}
}

func TestRight(t *testing.T) {
	t.Parallel()
	e := Right[string, int](42)
	if !e.IsRight() {
		t.Error("IsRight() = false; want true")
	}
	if e.IsLeft() {
		t.Error("IsLeft() = true; want false")
	}
	r, ok := e.RightVal()
	if !ok || r != 42 {
		t.Errorf("RightVal() = (%d, %v); want (42, true)", r, ok)
	}
	_, ok = e.LeftVal()
	if ok {
		t.Error("LeftVal() ok = true; want false")
	}
}

func TestLeftOr(t *testing.T) {
	t.Parallel()
	l := Left[string, int]("val")
	if v := l.LeftOr("default"); v != "val" {
		t.Errorf("Left.LeftOr = %q; want val", v)
	}
	r := Right[string, int](42)
	if v := r.LeftOr("default"); v != "default" {
		t.Errorf("Right.LeftOr = %q; want default", v)
	}
}

func TestRightOr(t *testing.T) {
	t.Parallel()
	r := Right[string, int](42)
	if v := r.RightOr(0); v != 42 {
		t.Errorf("Right.RightOr = %d; want 42", v)
	}
	l := Left[string, int]("err")
	if v := l.RightOr(99); v != 99 {
		t.Errorf("Left.RightOr = %d; want 99", v)
	}
}

func TestEitherMarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		e    Either[string, int]
		want string
	}{
		{"left", Left[string, int]("error"), `{"left":"error"}`},
		{"right", Right[string, int](42), `{"right":42}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.e)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("Marshal = %s; want %s", data, tt.want)
			}
		})
	}
}

func TestEitherUnmarshalJSON(t *testing.T) {
	t.Parallel()
	var e Either[string, int]
	if err := json.Unmarshal([]byte(`{"right":42}`), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !e.IsRight() {
		t.Error("expected Right")
	}
	r, _ := e.RightVal()
	if r != 42 {
		t.Errorf("RightVal = %d; want 42", r)
	}

	if err := json.Unmarshal([]byte(`{"left":"err"}`), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !e.IsLeft() {
		t.Error("expected Left")
	}
	l, _ := e.LeftVal()
	if l != "err" {
		t.Errorf("LeftVal = %q; want err", l)
	}
}

func TestEitherUnmarshalJSONEmpty(t *testing.T) {
	t.Parallel()
	var e Either[string, int]
	if err := json.Unmarshal([]byte(`{}`), &e); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !e.IsLeft() {
		t.Error("expected Left for empty JSON")
	}
}

func TestEitherUnmarshalJSONError(t *testing.T) {
	t.Parallel()
	var e Either[int, int]
	// "left" should be int but we provide a non-deserializable value
	err := json.Unmarshal([]byte(`{"left": "not_an_int"}`), &e)
	if err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestEitherRoundTrip(t *testing.T) {
	t.Parallel()
	original := Right[string, int](100)
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded Either[string, int]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !decoded.IsRight() {
		t.Error("expected Right after roundtrip")
	}
	r, _ := decoded.RightVal()
	if r != 100 {
		t.Errorf("RightVal = %d; want 100", r)
	}
}

func BenchmarkEitherRight(b *testing.B) {
	for range b.N {
		e := Right[string, int](42)
		e.RightOr(0)
	}
}

func BenchmarkEitherMarshal(b *testing.B) {
	e := Right[string, int](42)
	for range b.N {
		_, _ = json.Marshal(e)
	}
}

func FuzzEitherJSON(f *testing.F) {
	f.Add(42)
	f.Add(0)
	f.Add(-1)
	f.Fuzz(func(t *testing.T, v int) {
		e := Right[string, int](v)
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var got Either[string, int]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		r, ok := got.RightVal()
		if !ok || r != v {
			t.Errorf("roundtrip %d failed", v)
		}
	})
}
