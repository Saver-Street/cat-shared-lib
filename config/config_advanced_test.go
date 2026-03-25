package config

import (
	"fmt"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

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

func TestMustAddr_Valid(t *testing.T) {
	t.Setenv("TEST_ADDR", "localhost:8080")
	got := MustAddr("TEST_ADDR")
	testkit.AssertEqual(t, got, "localhost:8080")
}

func TestMustAddr_PanicsOnMissing(t *testing.T) {
	testkit.AssertPanics(t, func() { MustAddr("MISSING_ADDR") })
}

func TestMustAddr_PanicsOnInvalid(t *testing.T) {
	t.Setenv("TEST_ADDR", "not-an-addr")
	testkit.AssertPanics(t, func() { MustAddr("TEST_ADDR") })
}

func TestMustStringSlice_Valid(t *testing.T) {
	t.Setenv("TEST_SLICE", "a,b,c")
	got := MustStringSlice("TEST_SLICE")
	testkit.AssertLen(t, got, 3)
	testkit.AssertEqual(t, got[0], "a")
}

func TestMustStringSlice_PanicsOnMissing(t *testing.T) {
	testkit.AssertPanics(t, func() { MustStringSlice("MISSING_SLICE") })
}

func TestMustStringSlice_PanicsOnEmpty(t *testing.T) {
	t.Setenv("TEST_SLICE", "")
	testkit.AssertPanics(t, func() { MustStringSlice("TEST_SLICE") })
}

func TestFloat64_Default(t *testing.T) {
	got := Float64("MISSING_FLOAT", 3.14)
	testkit.AssertEqual(t, got, 3.14)
}

func TestFloat64_Set(t *testing.T) {
	t.Setenv("TEST_FLOAT", "2.718")
	got := Float64("TEST_FLOAT", 0)
	testkit.AssertEqual(t, got, 2.718)
}

func TestFloat64_Invalid(t *testing.T) {
	t.Setenv("TEST_FLOAT", "not-a-number")
	got := Float64("TEST_FLOAT", 1.0)
	testkit.AssertEqual(t, got, 1.0)
}

func TestMustFloat64_Valid(t *testing.T) {
	t.Setenv("TEST_FLOAT", "42.5")
	got := MustFloat64("TEST_FLOAT")
	testkit.AssertEqual(t, got, 42.5)
}

func TestMustFloat64_PanicsOnMissing(t *testing.T) {
	testkit.AssertPanics(t, func() { MustFloat64("MISSING_FLOAT") })
}

func TestMustFloat64_PanicsOnInvalid(t *testing.T) {
	t.Setenv("TEST_FLOAT", "abc")
	testkit.AssertPanics(t, func() { MustFloat64("TEST_FLOAT") })
}

func TestFilePath_Valid(t *testing.T) {
	t.Setenv("TEST_FILE", "config.go")
	got, err := FilePath("TEST_FILE", "")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, "config.go")
}

func TestFilePath_Default(t *testing.T) {
	got, err := FilePath("MISSING_FILE", "config.go")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, "config.go")
}

func TestFilePath_NotExist(t *testing.T) {
	t.Setenv("TEST_FILE", "/no/such/file/xyz")
	_, err := FilePath("TEST_FILE", "")
	testkit.AssertTrue(t, err != nil)
	testkit.AssertContains(t, err.Error(), "does not exist")
}

func TestFilePath_IsDirectory(t *testing.T) {
	t.Setenv("TEST_FILE", ".")
	_, err := FilePath("TEST_FILE", "")
	testkit.AssertTrue(t, err != nil)
	testkit.AssertContains(t, err.Error(), "is a directory")
}

func TestFilePath_Empty(t *testing.T) {
	got, err := FilePath("MISSING_FILE", "")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, "")
}

func TestDirPath_Valid(t *testing.T) {
	t.Setenv("TEST_DIR", ".")
	got, err := DirPath("TEST_DIR", "")
	testkit.AssertNoError(t, err)
	testkit.AssertEqual(t, got, ".")
}

func TestDirPath_NotExist(t *testing.T) {
	t.Setenv("TEST_DIR", "/no/such/dir/xyz")
	_, err := DirPath("TEST_DIR", "")
	testkit.AssertTrue(t, err != nil)
	testkit.AssertContains(t, err.Error(), "does not exist")
}

func TestDirPath_IsFile(t *testing.T) {
	t.Setenv("TEST_DIR", "config.go")
	_, err := DirPath("TEST_DIR", "")
	testkit.AssertTrue(t, err != nil)
	testkit.AssertContains(t, err.Error(), "is not a directory")
}

func TestPrefix_String(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	p := Prefix("DB_")
	testkit.AssertEqual(t, p.String("HOST", ""), "localhost")
}

func TestPrefix_Int(t *testing.T) {
	t.Setenv("DB_PORT", "5432")
	p := Prefix("DB_")
	testkit.AssertEqual(t, p.Int("PORT", 0), 5432)
}

func TestPrefix_Bool(t *testing.T) {
	t.Setenv("DB_SSL", "true")
	p := Prefix("DB_")
	testkit.AssertEqual(t, p.Bool("SSL", false), true)
}

func TestPrefix_Duration(t *testing.T) {
	t.Setenv("DB_TIMEOUT", "30s")
	p := Prefix("DB_")
	testkit.AssertEqual(t, p.Duration("TIMEOUT", 0), 30*time.Second)
}

func TestPrefix_MustString(t *testing.T) {
	t.Setenv("DB_NAME", "mydb")
	p := Prefix("DB_")
	testkit.AssertEqual(t, p.MustString("NAME"), "mydb")
}

func TestPrefix_MustString_Panics(t *testing.T) {
	p := Prefix("DB_")
	testkit.AssertPanics(t, func() { p.MustString("MISSING_KEY") })
}

func TestSummary_Basic(t *testing.T) {
	t.Setenv("APP_HOST", "localhost")
	t.Setenv("APP_PORT", "8080")
	m := Summary("APP_HOST", "APP_PORT", "APP_MISSING")
	testkit.AssertEqual(t, m["APP_HOST"], "localhost")
	testkit.AssertEqual(t, m["APP_PORT"], "8080")
	testkit.AssertEqual(t, m["APP_MISSING"], "(unset)")
}

func TestSummary_MasksSecrets(t *testing.T) {
	t.Setenv("DB_PASSWORD", "s3cret")
	t.Setenv("API_TOKEN", "tok123")
	t.Setenv("SIGNING_KEY", "abc")
	t.Setenv("CLIENT_SECRET", "xyz")
	m := Summary("DB_PASSWORD", "API_TOKEN", "SIGNING_KEY", "CLIENT_SECRET")
	testkit.AssertEqual(t, m["DB_PASSWORD"], "****")
	testkit.AssertEqual(t, m["API_TOKEN"], "****")
	testkit.AssertEqual(t, m["SIGNING_KEY"], "****")
	testkit.AssertEqual(t, m["CLIENT_SECRET"], "****")
}
