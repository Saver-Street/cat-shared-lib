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
	testkit.AssertEqual(t, String("CONFIG_TEST_UNSET_STR", "fallback"), "fallback")
}

func TestString_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_STR", "hello")
	testkit.AssertEqual(t, String("CONFIG_TEST_STR", "fallback"), "hello")
}

func TestStringRequired_Missing(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_REQ_MISS")
	_, err := StringRequired("CONFIG_TEST_REQ_MISS")
	testkit.AssertError(t, err)
}

func TestStringRequired_Present(t *testing.T) {
	setEnv(t, "CONFIG_TEST_REQ", "value")
	v, err := StringRequired("CONFIG_TEST_REQ")
	testkit.RequireNoError(t, err)
	testkit.AssertEqual(t, v, "value")
}

func TestInt_Default(t *testing.T) {
	testkit.AssertEqual(t, Int("CONFIG_TEST_UNSET_INT", 42), 42)
}

func TestInt_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_INT", "99")
	testkit.AssertEqual(t, Int("CONFIG_TEST_INT", 42), 99)
}

func TestInt_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_INT_BAD", "abc")
	testkit.AssertEqual(t, Int("CONFIG_TEST_INT_BAD", 42), 42)
}

func TestBool_Default(t *testing.T) {
	testkit.AssertTrue(t, Bool("CONFIG_TEST_UNSET_BOOL", true))
}

func TestBool_True(t *testing.T) {
	for _, v := range []string{"true", "1", "yes", "TRUE", "Yes"} {
		setEnv(t, "CONFIG_TEST_BOOL", v)
		testkit.AssertTrue(t, Bool("CONFIG_TEST_BOOL", false))
	}
}

func TestBool_False(t *testing.T) {
	for _, v := range []string{"false", "0", "no", "FALSE", "No"} {
		setEnv(t, "CONFIG_TEST_BOOL", v)
		testkit.AssertFalse(t, Bool("CONFIG_TEST_BOOL", true))
	}
}

func TestBool_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_BOOL_BAD", "maybe")
	testkit.AssertTrue(t, Bool("CONFIG_TEST_BOOL_BAD", true))
}

func TestDuration_Default(t *testing.T) {
	testkit.AssertEqual(t, Duration("CONFIG_TEST_UNSET_DUR", 5*time.Second), 5*time.Second)
}

func TestDuration_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_DUR", "30s")
	testkit.AssertEqual(t, Duration("CONFIG_TEST_DUR", 5*time.Second), 30*time.Second)
}

func TestDuration_Invalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_DUR_BAD", "not-a-duration")
	testkit.AssertEqual(t, Duration("CONFIG_TEST_DUR_BAD", 5*time.Second), 5*time.Second)
}

func TestStringSlice_Default(t *testing.T) {
	def := []string{"a", "b"}
	testkit.AssertEqual(t, StringSlice("CONFIG_TEST_UNSET_SLICE", def), def)
}

func TestStringSlice_Set(t *testing.T) {
	setEnv(t, "CONFIG_TEST_SLICE", "x, y, z")
	testkit.AssertEqual(t, StringSlice("CONFIG_TEST_SLICE", nil), []string{"x", "y", "z"})
}

func TestStringSlice_EmptyParts(t *testing.T) {
	setEnv(t, "CONFIG_TEST_SLICE_EMPTY", ", ,")
	testkit.AssertEqual(t, StringSlice("CONFIG_TEST_SLICE_EMPTY", []string{"default"}), []string{"default"})
}

func TestMustString_Panics(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_MUST_MISS")
	testkit.AssertPanics(t, func() {
		MustString("CONFIG_TEST_MUST_MISS")
	})
}

func TestMustString_Returns(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST", "present")
	testkit.AssertEqual(t, MustString("CONFIG_TEST_MUST"), "present")
}

func TestValidate_AllPresent(t *testing.T) {
	setEnv(t, "CONFIG_V1", "a")
	setEnv(t, "CONFIG_V2", "b")
	testkit.AssertNoError(t, Validate("CONFIG_V1", "CONFIG_V2"))
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
	testkit.AssertNoError(t, Validate())
}

func TestMustInt_Success(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_INT", "42")
	testkit.AssertEqual(t, MustInt("CONFIG_TEST_MUST_INT"), 42)
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

func TestMustBool_True(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_BOOL", "true")
	testkit.AssertTrue(t, MustBool("CONFIG_TEST_MUST_BOOL"))
}

func TestMustBool_False(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_BOOL", "false")
	testkit.AssertFalse(t, MustBool("CONFIG_TEST_MUST_BOOL"))
}

func TestMustBool_Yes(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_BOOL", "YES")
	testkit.AssertTrue(t, MustBool("CONFIG_TEST_MUST_BOOL"))
}

func TestMustBool_PanicsMissing(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_MUST_BOOL_MISS")
	testkit.AssertPanics(t, func() {
		MustBool("CONFIG_TEST_MUST_BOOL_MISS")
	})
}

func TestMustBool_PanicsInvalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_BOOL_BAD", "maybe")
	testkit.AssertPanics(t, func() {
		MustBool("CONFIG_TEST_MUST_BOOL_BAD")
	})
}

func TestMustDuration_Valid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_DUR", "5s")
	testkit.AssertEqual(t, MustDuration("CONFIG_TEST_MUST_DUR"), 5*time.Second)
}

func TestMustDuration_PanicsMissing(t *testing.T) {
	os.Unsetenv("CONFIG_TEST_MUST_DUR_MISS")
	testkit.AssertPanics(t, func() {
		MustDuration("CONFIG_TEST_MUST_DUR_MISS")
	})
}

func TestMustDuration_PanicsInvalid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MUST_DUR_BAD", "not-a-duration")
	testkit.AssertPanics(t, func() {
		MustDuration("CONFIG_TEST_MUST_DUR_BAD")
	})
}

func TestStringMap_Valid(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MAP", "env=prod, region=us-east")
	got := StringMap("CONFIG_TEST_MAP", nil)
	testkit.AssertEqual(t, got["env"], "prod")
	testkit.AssertEqual(t, got["region"], "us-east")
	testkit.AssertLen(t, got, 2)
}

func TestStringMap_Default(t *testing.T) {
	def := map[string]string{"a": "b"}
	got := StringMap("CONFIG_TEST_MAP_UNSET", def)
	testkit.AssertEqual(t, got["a"], "b")
}

func TestStringMap_SkipsBadPairs(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MAP_BAD", "good=val, noequalhere, another=ok")
	got := StringMap("CONFIG_TEST_MAP_BAD", nil)
	testkit.AssertLen(t, got, 2)
	testkit.AssertEqual(t, got["good"], "val")
	testkit.AssertEqual(t, got["another"], "ok")
}

func TestStringMap_EmptyValue(t *testing.T) {
	setEnv(t, "CONFIG_TEST_MAP_EMPTY", "key=")
	got := StringMap("CONFIG_TEST_MAP_EMPTY", nil)
	testkit.AssertLen(t, got, 1)
	testkit.AssertEqual(t, got["key"], "")
}

func TestStringMap_AllInvalid(t *testing.T) {
	def := map[string]string{"fallback": "yes"}
	setEnv(t, "CONFIG_TEST_MAP_ALLINV", "nope, bad, broken")
	got := StringMap("CONFIG_TEST_MAP_ALLINV", def)
	testkit.AssertEqual(t, got["fallback"], "yes")
}
