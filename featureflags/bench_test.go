package featureflags

import "testing"

func newBenchManager() *Manager {
	m := NewManager(Config{Prefix: "FF_"})
	m.resolver = func(key string) string {
		switch key {
		case "FF_DARK_MODE":
			return "true"
		case "FF_MAX_ITEMS":
			return "100"
		case "FF_RATIO":
			return "0.75"
		case "FF_REGIONS":
			return "us,eu,ap"
		default:
			return ""
		}
	}
	m.Register("DARK_MODE", "false", "Dark mode toggle")
	m.Register("MAX_ITEMS", "10", "Max items per page")
	m.Register("RATIO", "0.5", "Sample ratio")
	m.Register("REGIONS", "us", "Active regions")
	return m
}

func BenchmarkDisabled(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.Disabled("DARK_MODE")
	}
}

func BenchmarkValue(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.Value("DARK_MODE")
	}
}

func BenchmarkIntValue(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.IntValue("MAX_ITEMS", 0)
	}
}

func BenchmarkFloat64Value(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.Float64Value("RATIO", 0)
	}
}

func BenchmarkListValue(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.ListValue("REGIONS", ",")
	}
}

func BenchmarkAll(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.All()
	}
}

func BenchmarkAllFlags(b *testing.B) {
	m := newBenchManager()
	for b.Loop() {
		m.AllFlags()
	}
}
