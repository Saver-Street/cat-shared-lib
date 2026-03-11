package featureflags

import (
	"log/slog"
	"testing"
)

func FuzzRegisterAndEnabled(f *testing.F) {
	f.Add("dark_mode", "false", "true")
	f.Add("", "", "")
	f.Add("feature-🚀", "1", "0")
	f.Add("my_flag", "yes", "no")
	f.Fuzz(func(t *testing.T, name, defaultVal, resolvedVal string) {
		mgr := NewManager(Config{
			Prefix: "TEST_",
			Logger: slog.Default(),
		})
		mgr.resolver = func(string) string { return resolvedVal }

		mgr.Register(name, defaultVal, "test flag")

		// Must not panic.
		_ = mgr.Enabled(name)
		_ = mgr.Disabled(name)
		_ = mgr.Value(name)
	})
}

func FuzzIntValue(f *testing.F) {
	f.Add("count", "42", 0)
	f.Add("count", "not-a-number", 10)
	f.Add("count", "", 5)
	f.Add("count", "-1", 0)
	f.Add("count", "999999999999999999999", 0)
	f.Fuzz(func(t *testing.T, name, resolvedVal string, fallback int) {
		mgr := NewManager(Config{
			Prefix: "TEST_",
			Logger: slog.Default(),
		})
		mgr.resolver = func(string) string { return resolvedVal }

		mgr.Register(name, "0", "test")

		// Must not panic.
		_ = mgr.IntValue(name, fallback)
	})
}

func FuzzFloat64Value(f *testing.F) {
	f.Add("rate", "3.14", 0.0)
	f.Add("rate", "not-a-float", 1.0)
	f.Add("rate", "", 0.5)
	f.Add("rate", "Inf", 0.0)
	f.Add("rate", "NaN", 0.0)
	f.Fuzz(func(t *testing.T, name, resolvedVal string, fallback float64) {
		mgr := NewManager(Config{
			Prefix: "TEST_",
			Logger: slog.Default(),
		})
		mgr.resolver = func(string) string { return resolvedVal }

		mgr.Register(name, "0", "test")

		// Must not panic.
		_ = mgr.Float64Value(name, fallback)
	})
}

func FuzzListValue(f *testing.F) {
	f.Add("tags", "a,b,c", ",")
	f.Add("tags", "", ",")
	f.Add("tags", "single", ",")
	f.Add("tags", "a::b::c", "::")
	f.Fuzz(func(t *testing.T, name, resolvedVal, sep string) {
		mgr := NewManager(Config{
			Prefix: "TEST_",
			Logger: slog.Default(),
		})
		mgr.resolver = func(string) string { return resolvedVal }

		mgr.Register(name, "", "test")

		// Must not panic.
		_ = mgr.ListValue(name, sep)
	})
}
