package convert_test

import (
"testing"
"time"

"github.com/Saver-Street/cat-shared-lib/convert"
)

// --- ToInt ---

func TestToInt_Valid(t *testing.T) {
if got := convert.ToInt("42", 0); got != 42 {
t.Errorf("ToInt(\"42\", 0) = %d; want 42", got)
}
}

func TestToInt_Negative(t *testing.T) {
if got := convert.ToInt("-7", 0); got != -7 {
t.Errorf("ToInt(\"-7\", 0) = %d; want -7", got)
}
}

func TestToInt_Invalid(t *testing.T) {
if got := convert.ToInt("abc", 99); got != 99 {
t.Errorf("ToInt(\"abc\", 99) = %d; want 99", got)
}
}

func TestToInt_Empty(t *testing.T) {
if got := convert.ToInt("", 5); got != 5 {
t.Errorf("ToInt empty = %d; want 5", got)
}
}

func TestToInt_Overflow(t *testing.T) {
if got := convert.ToInt("99999999999999999999", 1); got != 1 {
t.Errorf("ToInt overflow = %d; want 1", got)
}
}

// --- ToInt64 ---

func TestToInt64_Valid(t *testing.T) {
if got := convert.ToInt64("9223372036854775807", 0); got != 9223372036854775807 {
t.Error("ToInt64 max int64 failed")
}
}

func TestToInt64_Invalid(t *testing.T) {
if got := convert.ToInt64("nope", -1); got != -1 {
t.Errorf("ToInt64 invalid = %d; want -1", got)
}
}

// --- ToFloat64 ---

func TestToFloat64_Valid(t *testing.T) {
if got := convert.ToFloat64("3.14", 0); got != 3.14 {
t.Errorf("ToFloat64 valid = %f; want 3.14", got)
}
}

func TestToFloat64_Invalid(t *testing.T) {
if got := convert.ToFloat64("xyz", 1.5); got != 1.5 {
t.Errorf("ToFloat64 invalid = %f; want 1.5", got)
}
}

func TestToFloat64_Empty(t *testing.T) {
if got := convert.ToFloat64("", 2.0); got != 2.0 {
t.Errorf("ToFloat64 empty = %f; want 2.0", got)
}
}

// --- ToBool ---

func TestToBool_TrueValues(t *testing.T) {
for _, s := range []string{"true", "TRUE", "1", "yes", "YES", "on", "ON"} {
if got := convert.ToBool(s, false); !got {
t.Errorf("ToBool(%q, false) = false; want true", s)
}
}
}

func TestToBool_FalseValues(t *testing.T) {
for _, s := range []string{"false", "FALSE", "0", "no", "NO", "off", "OFF"} {
if got := convert.ToBool(s, true); got {
t.Errorf("ToBool(%q, true) = true; want false", s)
}
}
}

func TestToBool_Invalid(t *testing.T) {
if got := convert.ToBool("maybe", true); !got {
t.Error("ToBool invalid should return fallback true")
}
}

func TestToBool_Empty(t *testing.T) {
if got := convert.ToBool("", false); got {
t.Error("ToBool empty should return fallback false")
}
}

func TestToBool_Whitespace(t *testing.T) {
if got := convert.ToBool("  true  ", false); !got {
t.Error("ToBool with whitespace should return true")
}
}

// --- ToString ---

func TestToString_String(t *testing.T) {
if got := convert.ToString("hello"); got != "hello" {
t.Errorf("ToString string = %q; want hello", got)
}
}

func TestToString_Int(t *testing.T) {
if got := convert.ToString(42); got != "42" {
t.Errorf("ToString int = %q; want 42", got)
}
}

func TestToString_Int64(t *testing.T) {
if got := convert.ToString(int64(100)); got != "100" {
t.Errorf("ToString int64 = %q; want 100", got)
}
}

func TestToString_Float64(t *testing.T) {
if got := convert.ToString(3.14); got != "3.14" {
t.Errorf("ToString float64 = %q; want 3.14", got)
}
}

func TestToString_Bool(t *testing.T) {
if got := convert.ToString(true); got != "true" {
t.Errorf("ToString bool = %q; want true", got)
}
}

func TestToString_Nil(t *testing.T) {
if got := convert.ToString(nil); got != "" {
t.Errorf("ToString nil = %q; want empty", got)
}
}

func TestToString_Unicode(t *testing.T) {
if got := convert.ToString("\u65e5\u672c\u8a9e"); got != "\u65e5\u672c\u8a9e" {
t.Error("ToString unicode failed")
}
}

// --- ToDuration ---

func TestToDuration_Valid(t *testing.T) {
if got := convert.ToDuration("5s", 0); got != 5*time.Second {
t.Errorf("ToDuration valid = %v; want 5s", got)
}
}

func TestToDuration_Invalid(t *testing.T) {
fb := 10 * time.Second
if got := convert.ToDuration("bad", fb); got != fb {
t.Errorf("ToDuration invalid = %v; want 10s", got)
}
}

// --- ToUint ---

func TestToUint_Valid(t *testing.T) {
if got := convert.ToUint("100", 0); got != 100 {
t.Errorf("ToUint valid = %d; want 100", got)
}
}

func TestToUint_Negative(t *testing.T) {
if got := convert.ToUint("-1", 0); got != 0 {
t.Errorf("ToUint negative = %d; want 0", got)
}
}

// --- MustInt ---

func TestMustInt_Valid(t *testing.T) {
if got := convert.MustInt("42"); got != 42 {
t.Errorf("MustInt valid = %d; want 42", got)
}
}

func TestMustInt_Panic(t *testing.T) {
defer func() {
if r := recover(); r == nil {
t.Error("MustInt should panic on bad input")
}
}()
convert.MustInt("bad")
}

// --- MustFloat64 ---

func TestMustFloat64_Valid(t *testing.T) {
if got := convert.MustFloat64("2.5"); got != 2.5 {
t.Errorf("MustFloat64 valid = %f; want 2.5", got)
}
}

func TestMustFloat64_Panic(t *testing.T) {
defer func() {
if r := recover(); r == nil {
t.Error("MustFloat64 should panic on bad input")
}
}()
convert.MustFloat64("bad")
}

// --- Ptr ---

func TestPtr_Int(t *testing.T) {
p := convert.Ptr(42)
if *p != 42 {
t.Errorf("Ptr(42) = %d; want 42", *p)
}
}

func TestPtr_String(t *testing.T) {
p := convert.Ptr("hello")
if *p != "hello" {
t.Errorf("Ptr string = %q; want hello", *p)
}
}

// --- Deref ---

func TestDeref_NonNil(t *testing.T) {
v := 42
if got := convert.Deref(&v); got != 42 {
t.Errorf("Deref non-nil = %d; want 42", got)
}
}

func TestDeref_Nil(t *testing.T) {
var p *int
if got := convert.Deref(p); got != 0 {
t.Errorf("Deref nil = %d; want 0", got)
}
}

// --- DerefOr ---

func TestDerefOr_NonNil(t *testing.T) {
v := 42
if got := convert.DerefOr(&v, 99); got != 42 {
t.Errorf("DerefOr non-nil = %d; want 42", got)
}
}

func TestDerefOr_Nil(t *testing.T) {
var p *string
if got := convert.DerefOr(p, "default"); got != "default" {
t.Errorf("DerefOr nil = %q; want default", got)
}
}

// --- Benchmarks ---

func BenchmarkToInt(b *testing.B) {
for b.Loop() {
convert.ToInt("12345", 0)
}
}

func BenchmarkToString(b *testing.B) {
for b.Loop() {
convert.ToString(12345)
}
}

func BenchmarkPtr(b *testing.B) {
for b.Loop() {
convert.Ptr(42)
}
}
