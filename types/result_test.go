package types

import (
	"errors"
	"strconv"
	"testing"
)

func TestOK(t *testing.T) {
	r := OK(42)
	if !r.IsOK() {
		t.Fatal("expected IsOK")
	}
	if r.IsErr() {
		t.Fatal("unexpected IsErr")
	}
	if r.Value() != 42 {
		t.Fatalf("value = %d, want 42", r.Value())
	}
	if r.Err() != nil {
		t.Fatalf("err = %v, want nil", r.Err())
	}
}

func TestFail(t *testing.T) {
	e := errors.New("oops")
	r := Fail[int](e)
	if r.IsOK() {
		t.Fatal("unexpected IsOK")
	}
	if !r.IsErr() {
		t.Fatal("expected IsErr")
	}
	if !errors.Is(r.Err(), e) {
		t.Fatalf("err = %v, want %v", r.Err(), e)
	}
	if r.Value() != 0 {
		t.Fatalf("value = %d, want 0", r.Value())
	}
}

func TestFromPair_OK(t *testing.T) {
	r := FromPair(strconv.Atoi("42"))
	if !r.IsOK() {
		t.Fatal("expected IsOK")
	}
	if r.Value() != 42 {
		t.Fatalf("value = %d, want 42", r.Value())
	}
}

func TestFromPair_Err(t *testing.T) {
	r := FromPair(strconv.Atoi("bad"))
	if !r.IsErr() {
		t.Fatal("expected IsErr")
	}
	if r.Err() == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestUnwrap(t *testing.T) {
	v, err := OK(99).Unwrap()
	if err != nil || v != 99 {
		t.Fatalf("Unwrap() = (%d, %v), want (99, nil)", v, err)
	}

	e := errors.New("fail")
	v2, err2 := Fail[int](e).Unwrap()
	if !errors.Is(err2, e) || v2 != 0 {
		t.Fatalf("Unwrap() = (%d, %v), want (0, %v)", v2, err2, e)
	}
}

func TestOrElse(t *testing.T) {
	if got := OK(10).OrElse(99); got != 10 {
		t.Fatalf("OrElse = %d, want 10", got)
	}
	if got := Fail[int](errors.New("x")).OrElse(99); got != 99 {
		t.Fatalf("OrElse = %d, want 99", got)
	}
}

func TestMap_OK(t *testing.T) {
	r := Map(OK(5), func(v int) string {
		return strconv.Itoa(v)
	})
	if !r.IsOK() {
		t.Fatal("expected IsOK")
	}
	if r.Value() != "5" {
		t.Fatalf("value = %q, want %q", r.Value(), "5")
	}
}

func TestMap_Err(t *testing.T) {
	e := errors.New("fail")
	r := Map(Fail[int](e), func(v int) string {
		return strconv.Itoa(v)
	})
	if !r.IsErr() {
		t.Fatal("expected IsErr")
	}
	if !errors.Is(r.Err(), e) {
		t.Fatalf("err = %v, want %v", r.Err(), e)
	}
}

func TestFlatMap_OK(t *testing.T) {
	r := FlatMap(OK(10), func(v int) Result[string] {
		return OK(strconv.Itoa(v))
	})
	if !r.IsOK() || r.Value() != "10" {
		t.Fatalf("FlatMap = (%q, %v)", r.Value(), r.Err())
	}
}

func TestFlatMap_Err(t *testing.T) {
	e := errors.New("fail")
	r := FlatMap(Fail[int](e), func(v int) Result[string] {
		return OK("never")
	})
	if !r.IsErr() || !errors.Is(r.Err(), e) {
		t.Fatalf("FlatMap = (%q, %v)", r.Value(), r.Err())
	}
}

func TestFlatMap_InnerErr(t *testing.T) {
	inner := errors.New("inner")
	r := FlatMap(OK(5), func(v int) Result[string] {
		return Fail[string](inner)
	})
	if !r.IsErr() || !errors.Is(r.Err(), inner) {
		t.Fatalf("FlatMap = (%q, %v)", r.Value(), r.Err())
	}
}

func TestResult_String(t *testing.T) {
	r := OK("hello")
	if r.Value() != "hello" {
		t.Fatalf("value = %q", r.Value())
	}
}

func BenchmarkOK(b *testing.B) {
	for b.Loop() {
		OK(42)
	}
}

func BenchmarkMap(b *testing.B) {
	r := OK(42)
	for b.Loop() {
		Map(r, func(v int) string { return strconv.Itoa(v) })
	}
}

func BenchmarkFlatMap(b *testing.B) {
	r := OK(42)
	for b.Loop() {
		FlatMap(r, func(v int) Result[string] { return OK(strconv.Itoa(v)) })
	}
}

func FuzzFromPair(f *testing.F) {
	f.Add("42")
	f.Add("bad")
	f.Add("")
	f.Add("-1")

	f.Fuzz(func(t *testing.T, s string) {
		r := FromPair(strconv.Atoi(s))
		_ = r.IsOK()
		_ = r.IsErr()
		_ = r.Value()
		_ = r.Err()
	})
}
