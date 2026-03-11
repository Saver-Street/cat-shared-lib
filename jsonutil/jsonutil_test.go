package jsonutil_test

import (
"encoding/json"
"testing"

"github.com/Saver-Street/cat-shared-lib/jsonutil"
)

// --------------- Pretty ---------------

func TestPretty_ValidJSON(t *testing.T) {
in := []byte(`{"a":1,"b":2}`)
out, err := jsonutil.Pretty(in)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !json.Valid(out) {
t.Fatal("output is not valid JSON")
}
if len(out) <= len(in) {
t.Fatal("expected pretty output to be longer")
}
}

func TestPretty_InvalidJSON(t *testing.T) {
_, err := jsonutil.Pretty([]byte(`{bad`))
if err == nil {
t.Fatal("expected error for invalid JSON")
}
}

// --------------- Compact ---------------

func TestCompact_ValidJSON(t *testing.T) {
in := []byte("{\n  \"a\": 1,\n  \"b\": 2\n}")
out, err := jsonutil.Compact(in)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
expected := `{"a":1,"b":2}`
if string(out) != expected {
t.Fatalf("got %s, want %s", out, expected)
}
}

func TestCompact_AlreadyCompact(t *testing.T) {
in := []byte(`{"x":"y"}`)
out, err := jsonutil.Compact(in)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if string(out) != string(in) {
t.Fatalf("got %s, want %s", out, in)
}
}

func TestCompact_InvalidJSON(t *testing.T) {
_, err := jsonutil.Compact([]byte(`not json`))
if err == nil {
t.Fatal("expected error for invalid JSON")
}
}

// --------------- Valid ---------------

func TestValid_True(t *testing.T) {
if !jsonutil.Valid([]byte(`{"key":"value"}`)) {
t.Fatal("expected valid")
}
}

func TestValid_TrueArray(t *testing.T) {
if !jsonutil.Valid([]byte(`[1,2,3]`)) {
t.Fatal("expected valid")
}
}

func TestValid_False(t *testing.T) {
if jsonutil.Valid([]byte(`{bad}`)) {
t.Fatal("expected invalid")
}
}

func TestValid_Empty(t *testing.T) {
if jsonutil.Valid([]byte(``)) {
t.Fatal("expected invalid for empty")
}
}

// --------------- MustMarshal ---------------

func TestMustMarshal_Success(t *testing.T) {
data := jsonutil.MustMarshal(map[string]int{"a": 1})
if !json.Valid(data) {
t.Fatal("expected valid JSON")
}
}

func TestMustMarshal_Panic(t *testing.T) {
defer func() {
if r := recover(); r == nil {
t.Fatal("expected panic")
}
}()
jsonutil.MustMarshal(make(chan int))
}

// --------------- MustUnmarshal ---------------

func TestMustUnmarshal_Success(t *testing.T) {
var m map[string]int
jsonutil.MustUnmarshal([]byte(`{"a":1}`), &m)
if m["a"] != 1 {
t.Fatalf("got %d, want 1", m["a"])
}
}

func TestMustUnmarshal_Panic(t *testing.T) {
defer func() {
if r := recover(); r == nil {
t.Fatal("expected panic")
}
}()
var m map[string]int
jsonutil.MustUnmarshal([]byte(`not json`), &m)
}

// --------------- Map ---------------

func TestMap_ValidObject(t *testing.T) {
m, err := jsonutil.Map([]byte(`{"name":"test","count":3}`))
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if m["name"] != "test" {
t.Fatalf("got %v, want test", m["name"])
}
}

func TestMap_InvalidJSON(t *testing.T) {
_, err := jsonutil.Map([]byte(`not json`))
if err == nil {
t.Fatal("expected error")
}
}

func TestMap_Array(t *testing.T) {
_, err := jsonutil.Map([]byte(`[1,2,3]`))
if err == nil {
t.Fatal("expected error for array input")
}
}

// --------------- Merge ---------------

func TestMerge_Simple(t *testing.T) {
a := []byte(`{"a":1}`)
b := []byte(`{"b":2}`)
out, err := jsonutil.Merge(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m, _ := jsonutil.Map(out)
if m["a"] != float64(1) || m["b"] != float64(2) {
t.Fatalf("unexpected merge result: %s", out)
}
}

func TestMerge_Override(t *testing.T) {
a := []byte(`{"x":1}`)
b := []byte(`{"x":2}`)
out, err := jsonutil.Merge(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m, _ := jsonutil.Map(out)
if m["x"] != float64(2) {
t.Fatalf("expected x=2, got %v", m["x"])
}
}

func TestMerge_Deep(t *testing.T) {
a := []byte(`{"nested":{"a":1,"b":2}}`)
b := []byte(`{"nested":{"b":3,"c":4}}`)
out, err := jsonutil.Merge(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
m, _ := jsonutil.Map(out)
nested := m["nested"].(map[string]any)
if nested["a"] != float64(1) || nested["b"] != float64(3) || nested["c"] != float64(4) {
t.Fatalf("unexpected deep merge: %s", out)
}
}

func TestMerge_InvalidA(t *testing.T) {
_, err := jsonutil.Merge([]byte(`bad`), []byte(`{}`))
if err == nil {
t.Fatal("expected error")
}
}

func TestMerge_InvalidB(t *testing.T) {
_, err := jsonutil.Merge([]byte(`{}`), []byte(`bad`))
if err == nil {
t.Fatal("expected error")
}
}

// --------------- GetPath ---------------

func TestGetPath_TopLevel(t *testing.T) {
v, err := jsonutil.GetPath([]byte(`{"name":"alice"}`), "name")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v != "alice" {
t.Fatalf("got %v, want alice", v)
}
}

func TestGetPath_Nested(t *testing.T) {
data := []byte(`{"user":{"profile":{"age":30}}}`)
v, err := jsonutil.GetPath(data, "user.profile.age")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v != float64(30) {
t.Fatalf("got %v, want 30", v)
}
}

func TestGetPath_MissingPath(t *testing.T) {
v, err := jsonutil.GetPath([]byte(`{"a":1}`), "b.c")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v != nil {
t.Fatalf("expected nil, got %v", v)
}
}

func TestGetPath_InvalidJSON(t *testing.T) {
_, err := jsonutil.GetPath([]byte(`bad`), "x")
if err == nil {
t.Fatal("expected error")
}
}

func TestGetPath_EmptyPath(t *testing.T) {
v, err := jsonutil.GetPath([]byte(`{"a":1}`), "")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if v == nil {
t.Fatal("expected non-nil for empty path")
}
}

// --------------- Equal ---------------

func TestEqual_SameContentDifferentWhitespace(t *testing.T) {
a := []byte(`{"a": 1, "b": 2}`)
b := []byte(`{"a":1,"b":2}`)
eq, err := jsonutil.Equal(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !eq {
t.Fatal("expected equal")
}
}

func TestEqual_DifferentContent(t *testing.T) {
a := []byte(`{"a":1}`)
b := []byte(`{"a":2}`)
eq, err := jsonutil.Equal(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if eq {
t.Fatal("expected not equal")
}
}

func TestEqual_InvalidA(t *testing.T) {
_, err := jsonutil.Equal([]byte(`bad`), []byte(`{}`))
if err == nil {
t.Fatal("expected error")
}
}

func TestEqual_InvalidB(t *testing.T) {
_, err := jsonutil.Equal([]byte(`{}`), []byte(`bad`))
if err == nil {
t.Fatal("expected error")
}
}

func TestEqual_Arrays(t *testing.T) {
a := []byte(`[1, 2, 3]`)
b := []byte(`[1,2,3]`)
eq, err := jsonutil.Equal(a, b)
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if !eq {
t.Fatal("expected equal")
}
}

// --------------- Benchmarks ---------------

func BenchmarkPretty(b *testing.B) {
data := []byte(`{"name":"test","nested":{"a":1,"b":[1,2,3]},"c":"hello"}`)
for b.Loop() {
jsonutil.Pretty(data)
}
}

func BenchmarkMerge(b *testing.B) {
a := []byte(`{"a":1,"nested":{"x":1}}`)
bData := []byte(`{"b":2,"nested":{"y":2}}`)
for b.Loop() {
jsonutil.Merge(a, bData)
}
}

func BenchmarkGetPath(b *testing.B) {
data := []byte(`{"user":{"profile":{"name":"alice","age":30}}}`)
for b.Loop() {
jsonutil.GetPath(data, "user.profile.name")
}
}

// --------------- Fuzz ---------------

func FuzzCompactRoundTrip(f *testing.F) {
f.Add([]byte(`{"a":1}`))
f.Add([]byte(`[1,2,3]`))
f.Add([]byte(`"hello"`))
f.Add([]byte(`null`))
f.Fuzz(func(t *testing.T, data []byte) {
if !json.Valid(data) {
return
}
pretty, err := jsonutil.Pretty(data)
if err != nil {
t.Fatalf("Pretty failed on valid JSON: %v", err)
}
compact, err := jsonutil.Compact(pretty)
if err != nil {
t.Fatalf("Compact failed on Pretty output: %v", err)
}
if !json.Valid(compact) {
t.Fatal("Compact(Pretty(data)) produced invalid JSON")
}
})
}
