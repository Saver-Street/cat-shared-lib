package types

import (
	"encoding/json"
	"testing"
)

func TestSome(t *testing.T) {
	t.Parallel()
	o := Some(42)
	if !o.IsSome() {
		t.Error("IsSome() = false; want true")
	}
	if o.IsNone() {
		t.Error("IsNone() = true; want false")
	}
	if v := o.Unwrap(); v != 42 {
		t.Errorf("Unwrap() = %d; want 42", v)
	}
}

func TestNone(t *testing.T) {
	t.Parallel()
	o := None[string]()
	if o.IsSome() {
		t.Error("IsSome() = true; want false")
	}
	if !o.IsNone() {
		t.Error("IsNone() = false; want true")
	}
}

func TestUnwrapPanic(t *testing.T) {
	t.Parallel()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if r != "option: unwrap called on None" {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	None[int]().Unwrap()
}

func TestUnwrapOr(t *testing.T) {
	t.Parallel()
	if v := Some(10).UnwrapOr(99); v != 10 {
		t.Errorf("Some.UnwrapOr = %d; want 10", v)
	}
	if v := None[int]().UnwrapOr(99); v != 99 {
		t.Errorf("None.UnwrapOr = %d; want 99", v)
	}
}

func TestUnwrapOrZero(t *testing.T) {
	t.Parallel()
	if v := Some("hello").UnwrapOrZero(); v != "hello" {
		t.Errorf("Some.UnwrapOrZero = %q; want hello", v)
	}
	if v := None[int]().UnwrapOrZero(); v != 0 {
		t.Errorf("None.UnwrapOrZero = %d; want 0", v)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()
	v, ok := Some(42).Get()
	if !ok || v != 42 {
		t.Errorf("Get() = (%d, %v); want (42, true)", v, ok)
	}
	_, ok = None[int]().Get()
	if ok {
		t.Error("None.Get() ok = true; want false")
	}
}

func TestOptionMarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		opt  Option[int]
		want string
	}{
		{"some", Some(42), "42"},
		{"none", None[int](), "null"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			data, err := json.Marshal(tt.opt)
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			if string(data) != tt.want {
				t.Errorf("Marshal = %s; want %s", data, tt.want)
			}
		})
	}
}

func TestOptionUnmarshalJSON(t *testing.T) {
	t.Parallel()
	var o Option[int]
	if err := json.Unmarshal([]byte("42"), &o); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !o.IsSome() || o.Unwrap() != 42 {
		t.Errorf("Unmarshal(42) = %v; want Some(42)", o)
	}

	if err := json.Unmarshal([]byte("null"), &o); err != nil {
		t.Fatalf("Unmarshal(null): %v", err)
	}
	if !o.IsNone() {
		t.Error("Unmarshal(null) should be None")
	}
}

func TestOptionUnmarshalJSONError(t *testing.T) {
	t.Parallel()
	var o Option[int]
	err := json.Unmarshal([]byte(`"not a number"`), &o)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestOptionJSONStruct(t *testing.T) {
	t.Parallel()
	type item struct {
		Name  string      `json:"name"`
		Value Option[int] `json:"value"`
	}
	i := item{Name: "test", Value: Some(5)}
	data, err := json.Marshal(i)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var got item
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !got.Value.IsSome() || got.Value.Unwrap() != 5 {
		t.Errorf("roundtrip = %v; want Some(5)", got.Value)
	}
}

func BenchmarkOptionUnwrapOr(b *testing.B) {
	o := Some(42)
	for range b.N {
		o.UnwrapOr(0)
	}
}

func BenchmarkOptionMarshalJSON(b *testing.B) {
	o := Some(42)
	for range b.N {
		_, _ = json.Marshal(o)
	}
}

func FuzzOptionJSON(f *testing.F) {
	f.Add(42)
	f.Add(0)
	f.Add(-1)
	f.Fuzz(func(t *testing.T, v int) {
		o := Some(v)
		data, err := json.Marshal(o)
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var got Option[int]
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if !got.IsSome() || got.Unwrap() != v {
			t.Errorf("roundtrip %d failed", v)
		}
	})
}
