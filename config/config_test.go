package config

import (
	"os"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func setEnv(t *testing.T, key, val string) {
	t.Helper()
	t.Setenv(key, val)
}

func TestString_Default(t *testing.T) {
	if got := String("CONFIG_TEST_UNSET_STR", "fallback"); got != "fallback" {
		t.Errorf("got %q, want %q", got, "fallback")
	}
}

func TestString_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_STR", "hello")
	if got := String("CONFIG_TEST_STR", "fallback"); got != "hello" {
		t.Errorf("got %q, want %q", got, "hello")
	}
}

func TestStringRequired_Missing(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_REQ_MISS")
	_, err := StringRequired("CONFIG_TEST_REQ_MISS")
	if err == nil {
		t.Error("expected error for missing required var")
	}
}

func TestStringRequired_Present(t *testing.T) {
	setEnv(t, "CONFIG_TEST_REQ", "value")
	v, err := StringRequired("CONFIG_TEST_REQ")
	if err != nil {
		t.Fatal(err)
	}
	if v != "value" {
		t.Errorf("got %q, want %q", v, "value")
	}
}

func TestInt_Default(t *testing.T) {
	if got := Int("CONFIG_TEST_UNSET_INT", 42); got != 42 {
		t.Errorf("got %d, want 42", got)
	}
}

func TestInt_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_INT", "99")
	if got := Int("CONFIG_TEST_INT", 42); got != 99 {
		t.Errorf("got %d, want 99", got)
	}
}

func TestInt_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_INT_BAD", "abc")
	if got := Int("CONFIG_TEST_INT_BAD", 42); got != 42 {
		t.Errorf("got %d, want 42 (default for invalid)", got)
	}
}

func TestBool_Default(t *testing.T) {
	if got := Bool("CONFIG_TEST_UNSET_BOOL", true); !got {
		t.Error("got false, want true (default)")
	}
}

func TestBool_True(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "TRUE", "Yes"} {
		setEnv(t, "CONFIG_TEST_BOOL", v)
		if got := Bool("CONFIG_TEST_BOOL", false); !got {
			t.Errorf("Bool(%q) = false, want true", v)
		}
	}
}

func TestBool_False(t *testing.T) {
	for _, v := range []string{"false", "0", "no", "FALSE", "No"} {
		setEnv(t, "CONFIG_TEST_BOOL", v)
		if got := Bool("CONFIG_TEST_BOOL", true); got {
			t.Errorf("Bool(%q) = true, want false", v)
		}
	}
}

func TestBool_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_BOOL_BAD", "maybe")
	if got := Bool("CONFIG_TEST_BOOL_BAD", true); !got {
		t.Error("invalid bool should return default (true)")
	}
}

func TestDuration_Default(t *testing.T) {
	d := Duration("CONFIG_TEST_UNSET_DUR", 5*time.Second)
	if d != 5*time.Second {
		t.Errorf("got %v, want 5s", d)
	}
}

func TestDuration_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_DUR", "30s")
	d := Duration("CONFIG_TEST_DUR", 5*time.Second)
	if d != 30*time.Second {
		t.Errorf("got %v, want 30s", d)
	}
}

func TestDuration_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_DUR_BAD", "not-a-duration")
	d := Duration("CONFIG_TEST_DUR_BAD", 5*time.Second)
	if d != 5*time.Second {
		t.Errorf("got %v, want 5s (default for invalid)", d)
	}
}

func TestStringSlice_Default(t *testing.T) {
	def := []string{"a", "b"}
	got := StringSlice("CONFIG_TEST_UNSET_SLICE", def)
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("got %v, want %v", got, def)
	}
}

func TestStringSlice_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_SLICE", "x, y, z")
	got := StringSlice("CONFIG_TEST_SLICE", nil)
	if len(got) != 3 || got[0] != "x" || got[1] != "y" || got[2] != "z" {
		t.Errorf("got %v, want [x y z]", got)
	}
}

func TestStringSlice_EmptyParts(t *testing.T) {
	setEnv(t, "CONFIG_TEST_SLICE_EMPTY", ", ,")
	got := StringSlice("CONFIG_TEST_SLICE_EMPTY", []string{"default"})
	if len(got) != 1 || got[0] != "default" {
		t.Errorf("all-empty parts should return default, got %v", got)
	}
}

func TestMustString_Panics(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_MUST_MISS")
	testkit.AssertPanics(t, func() {
		MustString("CONFIG_TEST_MUST_MISS")
	})
}

func TestMustString_Returns(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST", "present")
	if got := MustString("CONFIG_TEST_MUST"); got != "present" {
		t.Errorf("got %q, want %q", got, "present")
	}
}

func TestValidate_AllPresent(t *testing.T) {
	setEnv(t, "CONFIG_V1", "a")
	setEnv(t, "CONFIG_V2", "b")
	if err := Validate("CONFIG_V1", "CONFIG_V2"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_SomeMissing(t *testing.T) {
	setEnv(t, "CONFIG_V3", "a")
	os.Unsetenv("CONFIG_V4")
	os.Unsetenv("CONFIG_V5")
	err := Validate("CONFIG_V3", "CONFIG_V4", "CONFIG_V5")
	if err == nil {
		t.Fatal("expected error for missing vars")
	}
	testkit.AssertErrorContains(t, err, "CONFIG_V4")
	testkit.AssertErrorContains(t, err, "CONFIG_V5")
}

func TestValidate_NoneRequired(t *testing.T) {
	if err := Validate(); err != nil {
		t.Errorf("empty keys should pass: %v", err)
	}
}

func TestMustInt_Success(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_INT", "42")
	if got := MustInt("CONFIG_TEST_MUST_INT"); got != 42 {
		t.Errorf("MustInt = %d, want 42", got)
	}
}

func TestMustInt_PanicsMissing(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_MUST_INT_MISS")
	testkit.AssertPanics(t, func() {
		MustInt("CONFIG_TEST_MUST_INT_MISS")
	})
}

func TestMustInt_PanicsInvalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_INT_BAD", "not-a-number")
	testkit.AssertPanics(t, func() {
		MustInt("CONFIG_TEST_MUST_INT_BAD")
	})
}

func BenchmarkString(b *testing.B) {
	os.Setenv("CONFIG_BENCH", "value")
	defer os.Unsetenv("CONFIG_BENCH")
	for b.Loop() {
		String("CONFIG_BENCH", "default")
	}
}

func BenchmarkInt(b *testing.B) {
	os.Setenv("CONFIG_BENCH_INT", "42")
	defer os.Unsetenv("CONFIG_BENCH_INT")
	for b.Loop() {
		Int("CONFIG_BENCH_INT", 0)
	}
}
