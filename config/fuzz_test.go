package config

import (
	"testing"
)

func FuzzBytes(f *testing.F) {
	f.Add("0")
	f.Add("1024")
	f.Add("10KB")
	f.Add("5MB")
	f.Add("2GB")
	f.Add("1TB")
	f.Add("100B")
	f.Add("")
	f.Add("abc")
	f.Add("-1")
	f.Add("999999999999999999TB")
	f.Add("  42 KB  ")
	f.Add("0MB")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_BYTES", val)
		_, _ = Bytes("FUZZ_BYTES", 0) // must not panic
	})
}

func FuzzPort(f *testing.F) {
	f.Add("8080")
	f.Add("0")
	f.Add("65535")
	f.Add("65536")
	f.Add("-1")
	f.Add("")
	f.Add("abc")
	f.Add("443")
	f.Add("99999")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_PORT", val)
		_, _ = Port("FUZZ_PORT", 0) // must not panic
	})
}

func FuzzAddr(f *testing.F) {
	f.Add("localhost:8080")
	f.Add("0.0.0.0:443")
	f.Add(":3000")
	f.Add("")
	f.Add("no-port")
	f.Add("[::1]:8080")
	f.Add("host:99999")
	f.Add("host:abc")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_ADDR", val)
		_, _ = Addr("FUZZ_ADDR", "") // must not panic
	})
}

func FuzzURL(f *testing.F) {
	f.Add("https://example.com")
	f.Add("http://localhost:8080/path")
	f.Add("ftp://invalid.scheme")
	f.Add("")
	f.Add("not-a-url")
	f.Add("https://")
	f.Add("://missing-scheme")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_URL", val)
		_, _ = URL("FUZZ_URL", "") // must not panic
	})
}

func FuzzEnum(f *testing.F) {
	f.Add("debug")
	f.Add("info")
	f.Add("warn")
	f.Add("")
	f.Add("UNKNOWN")
	f.Add("  debug  ")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_ENUM", val)
		_, _ = Enum("FUZZ_ENUM", "", []string{"debug", "info", "warn", "error"}) // must not panic
	})
}

func FuzzStringMap(f *testing.F) {
	f.Add("key=val")
	f.Add("a=1,b=2,c=3")
	f.Add("")
	f.Add("no-equals")
	f.Add("=empty-key")
	f.Add("k=")
	f.Add(",,,")
	f.Add("  k = v , k2 = v2  ")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_MAP", val)
		_ = StringMap("FUZZ_MAP", nil) // must not panic
	})
}

func FuzzBool(f *testing.F) {
	f.Add("true")
	f.Add("false")
	f.Add("1")
	f.Add("0")
	f.Add("yes")
	f.Add("no")
	f.Add("")
	f.Add("TRUE")
	f.Add("maybe")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_BOOL", val)
		_ = Bool("FUZZ_BOOL", false) // must not panic
	})
}

func FuzzFloat64(f *testing.F) {
	f.Add("3.14")
	f.Add("-1.5")
	f.Add("0")
	f.Add("")
	f.Add("NaN")
	f.Add("Inf")
	f.Add("-Inf")
	f.Add("abc")
	f.Add("1e308")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_FLOAT", val)
		_ = Float64("FUZZ_FLOAT", 0) // must not panic
	})
}

func FuzzFeatureEnabled(f *testing.F) {
	f.Add("1")
	f.Add("true")
	f.Add("yes")
	f.Add("on")
	f.Add("0")
	f.Add("false")
	f.Add("")
	f.Add("OFF")
	f.Add("random")

	f.Fuzz(func(t *testing.T, val string) {
		t.Setenv("FUZZ_FEAT", val)
		_ = FeatureEnabled("FUZZ_FEAT") // must not panic
	})
}
