package errutil_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/errutil"
)

// ── Combine ────────────────────────────────────────────────────────

func TestCombine_AllNil(t *testing.T) {
	if err := errutil.Combine(nil, nil, nil); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestCombine_OneError(t *testing.T) {
	original := errors.New("fail")
	got := errutil.Combine(nil, original, nil)
	if got != original {
		t.Fatalf("expected original error, got %v", got)
	}
}

func TestCombine_MultipleErrors(t *testing.T) {
	e1 := errors.New("a")
	e2 := errors.New("b")
	got := errutil.Combine(nil, e1, nil, e2)
	if got == nil {
		t.Fatal("expected non-nil error")
	}
	msg := got.Error()
	if !strings.Contains(msg, "a") || !strings.Contains(msg, "b") {
		t.Fatalf("expected both messages, got %q", msg)
	}
}

func TestCombine_Empty(t *testing.T) {
	if err := errutil.Combine(); err != nil {
		t.Fatalf("expected nil for empty call, got %v", err)
	}
}

// ── Must ───────────────────────────────────────────────────────────

func TestMust_Success(t *testing.T) {
	v := errutil.Must(42, nil)
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

func TestMust_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if s, ok := r.(string); !ok || !strings.Contains(s, "errutil.Must") {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	errutil.Must(0, errors.New("boom"))
}

// ── MustOK ────────────────────────────────────────────────────────

func TestMustOK_Success(t *testing.T) {
	errutil.MustOK(nil) // should not panic
}

func TestMustOK_Panic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic")
		}
		if s, ok := r.(string); !ok || !strings.Contains(s, "errutil.MustOK") {
			t.Fatalf("unexpected panic value: %v", r)
		}
	}()
	errutil.MustOK(errors.New("boom"))
}

// ── Ignore ────────────────────────────────────────────────────────

func TestIgnore_ReturnsValue(t *testing.T) {
	v := errutil.Ignore("hello", errors.New("ignored"))
	if v != "hello" {
		t.Fatalf("expected hello, got %s", v)
	}
}

func TestIgnore_NilError(t *testing.T) {
	v := errutil.Ignore(99, nil)
	if v != 99 {
		t.Fatalf("expected 99, got %d", v)
	}
}

// ── Is ─────────────────────────────────────────────────────────────

func TestIs_Match(t *testing.T) {
	sentinel := errors.New("sentinel")
	wrapped := fmt.Errorf("context: %w", sentinel)
	if !errutil.Is(wrapped, sentinel) {
		t.Fatal("expected Is to match")
	}
}

func TestIs_NoMatch(t *testing.T) {
	if errutil.Is(errors.New("a"), errors.New("b")) {
		t.Fatal("expected Is to not match")
	}
}

// ── As ─────────────────────────────────────────────────────────────

type testErr struct{ Code int }

func (e *testErr) Error() string { return fmt.Sprintf("code %d", e.Code) }

func TestAs_Match(t *testing.T) {
	original := &testErr{Code: 42}
	wrapped := fmt.Errorf("wrapped: %w", original)
	got, ok := errutil.As[*testErr](wrapped)
	if !ok {
		t.Fatal("expected As to match")
	}
	if got.Code != 42 {
		t.Fatalf("expected Code 42, got %d", got.Code)
	}
}

func TestAs_NoMatch(t *testing.T) {
	_, ok := errutil.As[*testErr](errors.New("plain"))
	if ok {
		t.Fatal("expected As to not match")
	}
}

// ── Wrap ───────────────────────────────────────────────────────────

func TestWrap_WithError(t *testing.T) {
	original := errors.New("fail")
	wrapped := errutil.Wrap(original, "context")
	if wrapped == nil {
		t.Fatal("expected non-nil")
	}
	if !strings.Contains(wrapped.Error(), "context: fail") {
		t.Fatalf("unexpected message: %s", wrapped.Error())
	}
	if !errors.Is(wrapped, original) {
		t.Fatal("expected unwrap to work")
	}
}

func TestWrap_NilError(t *testing.T) {
	if errutil.Wrap(nil, "context") != nil {
		t.Fatal("expected nil for nil error")
	}
}

// ── Wrapf ──────────────────────────────────────────────────────────

func TestWrapf_WithError(t *testing.T) {
	original := errors.New("fail")
	wrapped := errutil.Wrapf(original, "op %d", 1)
	if wrapped == nil {
		t.Fatal("expected non-nil")
	}
	if !strings.Contains(wrapped.Error(), "op 1: fail") {
		t.Fatalf("unexpected message: %s", wrapped.Error())
	}
	if !errors.Is(wrapped, original) {
		t.Fatal("expected unwrap to work")
	}
}

func TestWrapf_NilError(t *testing.T) {
	if errutil.Wrapf(nil, "op %d", 1) != nil {
		t.Fatal("expected nil for nil error")
	}
}

// ── New ────────────────────────────────────────────────────────────

func TestNew_FormattedError(t *testing.T) {
	err := errutil.New("code %d: %s", 404, "not found")
	expected := "code 404: not found"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

// ── Recover ────────────────────────────────────────────────────────

func TestRecover_NoPanic(t *testing.T) {
	err := errutil.Recover(func() {})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecover_PanicWithError(t *testing.T) {
	original := errors.New("boom")
	err := errutil.Recover(func() { panic(original) })
	if err != original {
		t.Fatalf("expected original error, got %v", err)
	}
}

func TestRecover_PanicWithString(t *testing.T) {
	err := errutil.Recover(func() { panic("oops") })
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if !strings.Contains(err.Error(), "oops") {
		t.Fatalf("expected oops in message, got %s", err.Error())
	}
}

// ── RecoverFunc ────────────────────────────────────────────────────

func TestRecoverFunc_NoPanic(t *testing.T) {
	v, err := errutil.RecoverFunc(func() int { return 7 })
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if v != 7 {
		t.Fatalf("expected 7, got %d", v)
	}
}

func TestRecoverFunc_Panic(t *testing.T) {
	v, err := errutil.RecoverFunc(func() int { panic("fail") })
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if v != 0 {
		t.Fatalf("expected zero value, got %d", v)
	}
}

// ── Messages ──────────────────────────────────────────────────────

func TestMessages_SingleError(t *testing.T) {
	msgs := errutil.Messages(errors.New("single"))
	if len(msgs) != 1 || msgs[0] != "single" {
		t.Fatalf("unexpected: %v", msgs)
	}
}

func TestMessages_JoinedErrors(t *testing.T) {
	err := errors.Join(errors.New("a"), errors.New("b"), errors.New("c"))
	msgs := errutil.Messages(err)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d: %v", len(msgs), msgs)
	}
}

func TestMessages_Nil(t *testing.T) {
	msgs := errutil.Messages(nil)
	if msgs != nil {
		t.Fatalf("expected nil, got %v", msgs)
	}
}

// ── Benchmarks ────────────────────────────────────────────────────

func BenchmarkWrap(b *testing.B) {
	err := errors.New("base")
	for b.Loop() {
		_ = errutil.Wrap(err, "context")
	}
}

func BenchmarkCombine(b *testing.B) {
	e1 := errors.New("a")
	e2 := errors.New("b")
	e3 := errors.New("c")
	for b.Loop() {
		_ = errutil.Combine(nil, e1, nil, e2, e3)
	}
}
