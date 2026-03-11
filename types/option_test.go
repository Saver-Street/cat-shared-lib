package types

import (
	"encoding/json"
	"testing"
)

func TestSome(t *testing.T) {
	opt := Some(42)
	if !opt.IsPresent() {
		t.Error("Some should be present")
	}
	if opt.IsAbsent() {
		t.Error("Some should not be absent")
	}
	v, ok := opt.Value()
	if !ok || v != 42 {
		t.Errorf("Value() = (%v, %v), want (42, true)", v, ok)
	}
}

func TestNone(t *testing.T) {
	opt := None[string]()
	if opt.IsPresent() {
		t.Error("None should not be present")
	}
	if !opt.IsAbsent() {
		t.Error("None should be absent")
	}
	v, ok := opt.Value()
	if ok || v != "" {
		t.Errorf("Value() = (%q, %v), want (\"\", false)", v, ok)
	}
}

func TestOption_ZeroValue(t *testing.T) {
	var opt Option[int]
	if opt.IsPresent() {
		t.Error("zero value should be absent")
	}
	v, ok := opt.Value()
	if ok || v != 0 {
		t.Errorf("Value() = (%v, %v), want (0, false)", v, ok)
	}
}

func TestSome_ZeroValue(t *testing.T) {
	opt := Some(0)
	if !opt.IsPresent() {
		t.Error("Some(0) should be present")
	}
	v, ok := opt.Value()
	if !ok || v != 0 {
		t.Errorf("Value() = (%v, %v), want (0, true)", v, ok)
	}
}

func TestSome_EmptyString(t *testing.T) {
	opt := Some("")
	if !opt.IsPresent() {
		t.Error("Some(\"\") should be present")
	}
	v, ok := opt.Value()
	if !ok || v != "" {
		t.Errorf("Value() = (%q, %v), want (\"\", true)", v, ok)
	}
}

func TestOption_ValueOr(t *testing.T) {
	tests := []struct {
		name     string
		opt      Option[string]
		fallback string
		want     string
	}{
		{"present", Some("hello"), "default", "hello"},
		{"absent", None[string](), "default", "default"},
		{"present empty", Some(""), "default", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.opt.ValueOr(tt.fallback); got != tt.want {
				t.Errorf("ValueOr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOption_ValueOrFunc(t *testing.T) {
	called := false
	fn := func() int {
		called = true
		return 99
	}

	got := Some(42).ValueOrFunc(fn)
	if got != 42 {
		t.Errorf("ValueOrFunc(present) = %d, want 42", got)
	}
	if called {
		t.Error("fn should not be called when present")
	}

	got = None[int]().ValueOrFunc(fn)
	if got != 99 {
		t.Errorf("ValueOrFunc(absent) = %d, want 99", got)
	}
	if !called {
		t.Error("fn should be called when absent")
	}
}

func TestOption_MarshalJSON_Present(t *testing.T) {
	opt := Some(42)
	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != "42" {
		t.Errorf("Marshal = %s, want 42", data)
	}
}

func TestOption_MarshalJSON_Absent(t *testing.T) {
	opt := None[int]()
	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal = %s, want null", data)
	}
}

func TestOption_MarshalJSON_String(t *testing.T) {
	opt := Some("hello")
	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"hello"` {
		t.Errorf("Marshal = %s, want %q", data, `"hello"`)
	}
}

func TestOption_UnmarshalJSON_Value(t *testing.T) {
	var opt Option[int]
	if err := json.Unmarshal([]byte("42"), &opt); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if !opt.IsPresent() {
		t.Error("should be present after unmarshal")
	}
	v, _ := opt.Value()
	if v != 42 {
		t.Errorf("Value() = %d, want 42", v)
	}
}

func TestOption_UnmarshalJSON_Null(t *testing.T) {
	var opt Option[int]
	if err := json.Unmarshal([]byte("null"), &opt); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if opt.IsPresent() {
		t.Error("should be absent after unmarshal null")
	}
}

func TestOption_UnmarshalJSON_InvalidType(t *testing.T) {
	var opt Option[int]
	err := json.Unmarshal([]byte(`"not a number"`), &opt)
	if err == nil {
		t.Error("expected error for invalid type")
	}
}

func TestOption_JSONRoundTrip_Struct(t *testing.T) {
	type Request struct {
		Name  Option[string] `json:"name"`
		Age   Option[int]    `json:"age"`
		Email Option[string] `json:"email"`
	}

	original := Request{
		Name:  Some("Alice"),
		Age:   None[int](),
		Email: Some(""),
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded Request
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if v, ok := decoded.Name.Value(); !ok || v != "Alice" {
		t.Errorf("Name = (%q, %v), want (\"Alice\", true)", v, ok)
	}
	if decoded.Age.IsPresent() {
		t.Error("Age should be absent after round-trip")
	}
	if v, ok := decoded.Email.Value(); !ok || v != "" {
		t.Errorf("Email = (%q, %v), want (\"\", true)", v, ok)
	}
}

func TestOption_JSONRoundTrip_Slice(t *testing.T) {
	opt := Some([]string{"a", "b", "c"})
	data, err := json.Marshal(opt)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `["a","b","c"]` {
		t.Errorf("Marshal = %s", data)
	}

	var decoded Option[[]string]
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	v, ok := decoded.Value()
	if !ok || len(v) != 3 {
		t.Errorf("Value() = (%v, %v), want ([a b c], true)", v, ok)
	}
}

func TestOption_JSONPartialPatch(t *testing.T) {
	type PatchRequest struct {
		Name  Option[string] `json:"name"`
		Email Option[string] `json:"email"`
	}

	raw := `{"name": "Bob"}`
	var req PatchRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if !req.Name.IsPresent() {
		t.Error("Name should be present")
	}
	if v, _ := req.Name.Value(); v != "Bob" {
		t.Errorf("Name = %q, want %q", v, "Bob")
	}
	if req.Email.IsPresent() {
		t.Error("Email should be absent (not in JSON)")
	}
}

func TestOption_JSONExplicitNull(t *testing.T) {
	type PatchRequest struct {
		Name Option[string] `json:"name"`
	}

	raw := `{"name": null}`
	var req PatchRequest
	if err := json.Unmarshal([]byte(raw), &req); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if req.Name.IsPresent() {
		t.Error("explicit null should produce absent Option")
	}
}

func TestOption_SomeWithPointer(t *testing.T) {
	s := "hello"
	opt := Some(&s)
	if !opt.IsPresent() {
		t.Error("should be present")
	}
	v, ok := opt.Value()
	if !ok || *v != "hello" {
		t.Errorf("Value() = (%v, %v), want (hello, true)", v, ok)
	}
}

func BenchmarkSome(b *testing.B) {
	for b.Loop() {
		_ = Some(42)
	}
}

func BenchmarkNone(b *testing.B) {
	for b.Loop() {
		_ = None[int]()
	}
}

func BenchmarkOption_ValueOr(b *testing.B) {
	opt := Some("hello")
	for b.Loop() {
		_ = opt.ValueOr("default")
	}
}

func BenchmarkOption_MarshalJSON(b *testing.B) {
	opt := Some(42)
	for b.Loop() {
		_, _ = json.Marshal(opt)
	}
}

func BenchmarkOption_UnmarshalJSON(b *testing.B) {
	data := []byte("42")
	for b.Loop() {
		var opt Option[int]
		_ = json.Unmarshal(data, &opt)
	}
}

func FuzzOption_UnmarshalJSON(f *testing.F) {
	f.Add([]byte("42"))
	f.Add([]byte("null"))
	f.Add([]byte(`"hello"`))
	f.Add([]byte(`true`))
	f.Add([]byte(`[1,2,3]`))
	f.Add([]byte(`{}`))
	f.Add([]byte(``))
	f.Add([]byte(`"\x00\xff"`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var opt Option[any]
		_ = json.Unmarshal(data, &opt) // must not panic
	})
}
