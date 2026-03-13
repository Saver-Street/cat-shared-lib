package cond_test

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/cond"
)

// --------------- Ternary ---------------

func TestTernary_True(t *testing.T) {
	if got := cond.Ternary(true, "yes", "no"); got != "yes" {
		t.Errorf("Ternary(true) = %q; want %q", got, "yes")
	}
}

func TestTernary_False(t *testing.T) {
	if got := cond.Ternary(false, "yes", "no"); got != "no" {
		t.Errorf("Ternary(false) = %q; want %q", got, "no")
	}
}

func TestTernary_Int(t *testing.T) {
	if got := cond.Ternary(true, 1, 2); got != 1 {
		t.Errorf("Ternary(true, 1, 2) = %d; want 1", got)
	}
}

func TestTernary_Struct(t *testing.T) {
	type point struct{ X, Y int }
	a, b := point{1, 2}, point{3, 4}
	if got := cond.Ternary(false, a, b); got != b {
		t.Errorf("Ternary(false, a, b) = %v; want %v", got, b)
	}
}

// --------------- Coalesce ---------------

func TestCoalesce_FirstNonZero(t *testing.T) {
	if got := cond.Coalesce(0, 0, 42, 7); got != 42 {
		t.Errorf("Coalesce = %d; want 42", got)
	}
}

func TestCoalesce_AllZero(t *testing.T) {
	if got := cond.Coalesce(0, 0, 0); got != 0 {
		t.Errorf("Coalesce = %d; want 0", got)
	}
}

func TestCoalesce_SingleValue(t *testing.T) {
	if got := cond.Coalesce(99); got != 99 {
		t.Errorf("Coalesce = %d; want 99", got)
	}
}

func TestCoalesce_Strings(t *testing.T) {
	if got := cond.Coalesce("", "", "hello"); got != "hello" {
		t.Errorf("Coalesce = %q; want %q", got, "hello")
	}
}

func TestCoalesce_NoArgs(t *testing.T) {
	if got := cond.Coalesce[int](); got != 0 {
		t.Errorf("Coalesce() = %d; want 0", got)
	}
}

// --------------- CoalesceFunc ---------------

func TestCoalesceFunc_LazyEval(t *testing.T) {
	calls := 0
	got := cond.CoalesceFunc(
		func() int { calls++; return 0 },
		func() int { calls++; return 5 },
		func() int { calls++; return 9 },
	)
	if got != 5 {
		t.Errorf("CoalesceFunc = %d; want 5", got)
	}
	if calls != 2 {
		t.Errorf("calls = %d; want 2 (lazy)", calls)
	}
}

func TestCoalesceFunc_AllZero(t *testing.T) {
	got := cond.CoalesceFunc(
		func() string { return "" },
		func() string { return "" },
	)
	if got != "" {
		t.Errorf("CoalesceFunc = %q; want empty", got)
	}
}

func TestCoalesceFunc_NoArgs(t *testing.T) {
	if got := cond.CoalesceFunc[int](); got != 0 {
		t.Errorf("CoalesceFunc() = %d; want 0", got)
	}
}

// --------------- Clamp ---------------

func TestClamp_InRange(t *testing.T) {
	if got := cond.Clamp(5, 1, 10); got != 5 {
		t.Errorf("Clamp = %d; want 5", got)
	}
}

func TestClamp_BelowMin(t *testing.T) {
	if got := cond.Clamp(-3, 0, 10); got != 0 {
		t.Errorf("Clamp = %d; want 0", got)
	}
}

func TestClamp_AboveMax(t *testing.T) {
	if got := cond.Clamp(99, 0, 10); got != 10 {
		t.Errorf("Clamp = %d; want 10", got)
	}
}

func TestClamp_AtLow(t *testing.T) {
	if got := cond.Clamp(0, 0, 10); got != 0 {
		t.Errorf("Clamp = %d; want 0", got)
	}
}

func TestClamp_AtHigh(t *testing.T) {
	if got := cond.Clamp(10, 0, 10); got != 10 {
		t.Errorf("Clamp = %d; want 10", got)
	}
}

func TestClamp_Float(t *testing.T) {
	if got := cond.Clamp(3.14, 0.0, 1.0); got != 1.0 {
		t.Errorf("Clamp = %f; want 1.0", got)
	}
}

// --------------- Min ---------------

func TestMin_MultipleValues(t *testing.T) {
	if got := cond.Min(3, 1, 4, 1, 5); got != 1 {
		t.Errorf("Min = %d; want 1", got)
	}
}

func TestMin_SingleValue(t *testing.T) {
	if got := cond.Min(42); got != 42 {
		t.Errorf("Min = %d; want 42", got)
	}
}

func TestMin_Negative(t *testing.T) {
	if got := cond.Min(-1, -5, -3); got != -5 {
		t.Errorf("Min = %d; want -5", got)
	}
}

func TestMin_PanicOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Min() did not panic")
		}
	}()
	cond.Min[int]()
}

// --------------- Max ---------------

func TestMax_MultipleValues(t *testing.T) {
	if got := cond.Max(3, 1, 4, 1, 5); got != 5 {
		t.Errorf("Max = %d; want 5", got)
	}
}

func TestMax_SingleValue(t *testing.T) {
	if got := cond.Max(42); got != 42 {
		t.Errorf("Max = %d; want 42", got)
	}
}

func TestMax_Negative(t *testing.T) {
	if got := cond.Max(-1, -5, -3); got != -1 {
		t.Errorf("Max = %d; want -1", got)
	}
}

func TestMax_PanicOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Max() did not panic")
		}
	}()
	cond.Max[int]()
}

// --------------- Zero ---------------

func TestZero_Int(t *testing.T) {
	if got := cond.Zero[int](); got != 0 {
		t.Errorf("Zero[int]() = %d; want 0", got)
	}
}

func TestZero_String(t *testing.T) {
	if got := cond.Zero[string](); got != "" {
		t.Errorf("Zero[string]() = %q; want empty", got)
	}
}

func TestZero_Struct(t *testing.T) {
	type S struct {
		A int
		B string
	}
	if got := cond.Zero[S](); got != (S{}) {
		t.Errorf("Zero[S]() = %v; want zero", got)
	}
}

// --------------- IsZero ---------------

func TestIsZero_ZeroInt(t *testing.T) {
	if !cond.IsZero(0) {
		t.Error("IsZero(0) = false; want true")
	}
}

func TestIsZero_NonZeroInt(t *testing.T) {
	if cond.IsZero(1) {
		t.Error("IsZero(1) = true; want false")
	}
}

func TestIsZero_EmptyString(t *testing.T) {
	if !cond.IsZero("") {
		t.Error("IsZero(\"\") = false; want true")
	}
}

func TestIsZero_NonEmptyString(t *testing.T) {
	if cond.IsZero("hi") {
		t.Error("IsZero(\"hi\") = true; want false")
	}
}

// --------------- Switch ---------------

func TestSwitch_FirstMatch(t *testing.T) {
	got := cond.Switch(
		cond.Case[string]{When: false, Then: "a"},
		cond.Case[string]{When: true, Then: "b"},
		cond.Case[string]{When: true, Then: "c"},
	)
	if got != "b" {
		t.Errorf("Switch = %q; want %q", got, "b")
	}
}

func TestSwitch_NoMatch(t *testing.T) {
	got := cond.Switch(
		cond.Case[int]{When: false, Then: 1},
		cond.Case[int]{When: false, Then: 2},
	)
	if got != 0 {
		t.Errorf("Switch = %d; want 0", got)
	}
}

func TestSwitch_MultipleTrue(t *testing.T) {
	got := cond.Switch(
		cond.Case[string]{When: true, Then: "first"},
		cond.Case[string]{When: true, Then: "second"},
	)
	if got != "first" {
		t.Errorf("Switch = %q; want %q", got, "first")
	}
}

// --------------- Benchmarks ---------------

func BenchmarkTernary(b *testing.B) {
	for b.Loop() {
		cond.Ternary(true, "yes", "no")
	}
}

func BenchmarkCoalesce(b *testing.B) {
	for b.Loop() {
		cond.Coalesce(0, 0, 0, 42)
	}
}

func BenchmarkClamp(b *testing.B) {
	for b.Loop() {
		cond.Clamp(50, 0, 100)
	}
}
