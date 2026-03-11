package config

import (
	"fmt"
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

func TestURL_Valid(t *testing.T) {
setEnv(t, "CONFIG_TEST_URL", "https://api.example.com/v1")
got, err := URL("CONFIG_TEST_URL", "")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, got, "https://api.example.com/v1")
}

func TestURL_Default(t *testing.T) {
got, err := URL("CONFIG_TEST_URL_UNSET", "https://default.example.com")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, got, "https://default.example.com")
}

func TestURL_BadScheme(t *testing.T) {
setEnv(t, "CONFIG_TEST_URL_BAD", "ftp://files.example.com")
_, err := URL("CONFIG_TEST_URL_BAD", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "http or https")
}

func TestURL_NoHost(t *testing.T) {
setEnv(t, "CONFIG_TEST_URL_NOHOST", "https://")
_, err := URL("CONFIG_TEST_URL_NOHOST", "")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "must have a host")
}

func TestInt64(t *testing.T) {
setEnv(t, "CONFIG_TEST_INT64", "9223372036854775807")
got := Int64("CONFIG_TEST_INT64", 0)
testkit.AssertEqual(t, got, int64(9223372036854775807))
}

func TestInt64_Default(t *testing.T) {
got := Int64("CONFIG_TEST_INT64_UNSET", 42)
testkit.AssertEqual(t, got, int64(42))
}

func TestInt64_Invalid(t *testing.T) {
setEnv(t, "CONFIG_TEST_INT64_BAD", "not_a_number")
got := Int64("CONFIG_TEST_INT64_BAD", 99)
testkit.AssertEqual(t, got, int64(99))
}

func TestMustInt64(t *testing.T) {
setEnv(t, "CONFIG_TEST_MUST_INT64", "1234567890123")
got := MustInt64("CONFIG_TEST_MUST_INT64")
testkit.AssertEqual(t, got, int64(1234567890123))
}

func TestMustInt64_Panics(t *testing.T) {
testkit.AssertPanics(t, func() {
MustInt64("CONFIG_TEST_MUST_INT64_UNSET")
})
}

func TestMustInt64_PanicsInvalid(t *testing.T) {
setEnv(t, "CONFIG_TEST_MUST_INT64_INV", "abc")
testkit.AssertPanics(t, func() {
MustInt64("CONFIG_TEST_MUST_INT64_INV")
})
}

func TestPort_Valid(t *testing.T) {
setEnv(t, "CONFIG_TEST_PORT", "8080")
got, err := Port("CONFIG_TEST_PORT", 3000)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, got, 8080)
}

func TestPort_Default(t *testing.T) {
got, err := Port("CONFIG_TEST_PORT_UNSET", 3000)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, got, 3000)
}

func TestPort_TooHigh(t *testing.T) {
setEnv(t, "CONFIG_TEST_PORT_HIGH", "99999")
_, err := Port("CONFIG_TEST_PORT_HIGH", 3000)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "between 1 and 65535")
}

func TestPort_Zero(t *testing.T) {
setEnv(t, "CONFIG_TEST_PORT_ZERO", "0")
_, err := Port("CONFIG_TEST_PORT_ZERO", 3000)
testkit.AssertError(t, err)
}

func TestPort_Invalid(t *testing.T) {
setEnv(t, "CONFIG_TEST_PORT_INV", "abc")
_, err := Port("CONFIG_TEST_PORT_INV", 3000)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "invalid port number")
}

func TestAddr_Default(t *testing.T) {
v, err := Addr("CFG_TEST_ADDR_UNSET", ":8080")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, ":8080")
}

func TestAddr_Valid(t *testing.T) {
t.Setenv("CFG_TEST_ADDR", "localhost:3000")
v, err := Addr("CFG_TEST_ADDR", ":8080")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, "localhost:3000")
}

func TestAddr_ValidIP(t *testing.T) {
t.Setenv("CFG_TEST_ADDR_IP", "0.0.0.0:443")
v, err := Addr("CFG_TEST_ADDR_IP", ":8080")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, "0.0.0.0:443")
}

func TestAddr_EmptyHost(t *testing.T) {
t.Setenv("CFG_TEST_ADDR_EH", ":9090")
v, err := Addr("CFG_TEST_ADDR_EH", ":8080")
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, ":9090")
}

func TestAddr_InvalidNoPort(t *testing.T) {
t.Setenv("CFG_TEST_ADDR_NP", "localhost")
_, err := Addr("CFG_TEST_ADDR_NP", ":8080")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "invalid address")
}

func TestAddr_InvalidPort(t *testing.T) {
t.Setenv("CFG_TEST_ADDR_BAD", "localhost:99999")
_, err := Addr("CFG_TEST_ADDR_BAD", ":8080")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "port must be between")
}

func TestStringSliceRequired_Set(t *testing.T) {
t.Setenv("CFG_TEST_SSR", "a, b, c")
got, err := StringSliceRequired("CFG_TEST_SSR")
testkit.AssertNoError(t, err)
testkit.AssertLen(t, got, 3)
testkit.AssertEqual(t, got[0], "a")
testkit.AssertEqual(t, got[1], "b")
testkit.AssertEqual(t, got[2], "c")
}

func TestStringSliceRequired_Unset(t *testing.T) {
_, err := StringSliceRequired("CFG_TEST_SSR_UNSET")
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "is required")
}

func TestStringSliceRequired_Empty(t *testing.T) {
t.Setenv("CFG_TEST_SSR_EMPTY", "")
_, err := StringSliceRequired("CFG_TEST_SSR_EMPTY")
testkit.AssertError(t, err)
}

func TestStringSliceRequired_OnlyCommas(t *testing.T) {
t.Setenv("CFG_TEST_SSR_COMMAS", ", , ,")
_, err := StringSliceRequired("CFG_TEST_SSR_COMMAS")
testkit.AssertError(t, err)
}

func TestBytes_Default(t *testing.T) {
v, err := Bytes("CFG_TEST_BYTES_UNSET", 1024)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(1024))
}

func TestBytes_Plain(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_PLAIN", "4096")
v, err := Bytes("CFG_TEST_BYTES_PLAIN", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(4096))
}

func TestBytes_KB(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_KB", "512KB")
v, err := Bytes("CFG_TEST_BYTES_KB", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(512*1024))
}

func TestBytes_MB(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_MB", "64MB")
v, err := Bytes("CFG_TEST_BYTES_MB", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(64*1024*1024))
}

func TestBytes_GB(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_GB", "2GB")
v, err := Bytes("CFG_TEST_BYTES_GB", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(2*1024*1024*1024))
}

func TestBytes_B(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_B", "100B")
v, err := Bytes("CFG_TEST_BYTES_B", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(100))
}

func TestBytes_Invalid(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_INV", "abcMB")
_, err := Bytes("CFG_TEST_BYTES_INV", 0)
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "invalid byte size")
}

func TestBytes_Lowercase(t *testing.T) {
t.Setenv("CFG_TEST_BYTES_LC", "10mb")
v, err := Bytes("CFG_TEST_BYTES_LC", 0)
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, int64(10*1024*1024))
}

func TestEnum_Default(t *testing.T) {
v, err := Enum("TEST_ENUM_UNSET", "info", []string{"debug", "info", "warn", "error"})
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, "info")
}

func TestEnum_Valid(t *testing.T) {
t.Setenv("TEST_ENUM", "warn")
v, err := Enum("TEST_ENUM", "info", []string{"debug", "info", "warn", "error"})
testkit.AssertNoError(t, err)
testkit.AssertEqual(t, v, "warn")
}

func TestEnum_Invalid(t *testing.T) {
t.Setenv("TEST_ENUM", "trace")
_, err := Enum("TEST_ENUM", "info", []string{"debug", "info", "warn", "error"})
testkit.AssertError(t, err)
testkit.AssertContains(t, err.Error(), "trace")
testkit.AssertContains(t, err.Error(), "not one of")
}

func TestMustEnum_Valid(t *testing.T) {
t.Setenv("TEST_MUST_ENUM", "warn")
v := MustEnum("TEST_MUST_ENUM", []string{"debug", "info", "warn", "error"})
testkit.AssertEqual(t, v, "warn")
}

func TestMustEnum_Missing(t *testing.T) {
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "required")
}()
MustEnum("TEST_MUST_ENUM_UNSET", []string{"a", "b"})
}

func TestMustEnum_Invalid(t *testing.T) {
t.Setenv("TEST_MUST_ENUM", "trace")
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "not one of")
}()
MustEnum("TEST_MUST_ENUM", []string{"debug", "info", "warn"})
}

func TestMustURL_Valid(t *testing.T) {
t.Setenv("TEST_MUST_URL", "https://api.example.com")
v := MustURL("TEST_MUST_URL")
testkit.AssertEqual(t, v, "https://api.example.com")
}

func TestMustURL_Missing(t *testing.T) {
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "required")
}()
MustURL("TEST_MUST_URL_UNSET")
}

func TestMustURL_Invalid(t *testing.T) {
t.Setenv("TEST_MUST_URL", "ftp://example.com")
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "http or https")
}()
MustURL("TEST_MUST_URL")
}

func TestMustPort_Valid(t *testing.T) {
t.Setenv("TEST_MUST_PORT", "8080")
v := MustPort("TEST_MUST_PORT")
testkit.AssertEqual(t, v, 8080)
}

func TestMustPort_Missing(t *testing.T) {
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "required")
}()
MustPort("TEST_MUST_PORT_UNSET")
}

func TestMustPort_Invalid(t *testing.T) {
t.Setenv("TEST_MUST_PORT", "99999")
defer func() {
r := recover()
testkit.AssertNotNil(t, r)
testkit.AssertContains(t, fmt.Sprint(r), "between 1 and 65535")
}()
MustPort("TEST_MUST_PORT")
}
