package featureflags

import (
	"bytes"
	"log/slog"
	"testing"
)

func newTestManager() *Manager {
	env := map[string]string{}
	m := NewManager(Config{Prefix: "FF_"})
	m.resolver = func(key string) string { return env[key] }
	return m
}

func newTestManagerWithEnv(env map[string]string) *Manager {
	m := NewManager(Config{Prefix: "FF_"})
	m.resolver = func(key string) string { return env[key] }
	return m
}

func TestNewManager_Defaults(t *testing.T) {
	m := NewManager(Config{})
	if m.prefix != "FEATURE_" {
		t.Errorf("expected default prefix FEATURE_, got %q", m.prefix)
	}
	if m.logger == nil {
		t.Error("expected non-nil logger")
	}
}

func TestRegister_And_Enabled(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_NEW_UI": "true"})
	m.Register("new_ui", "false", "Enable new UI")

	if !m.Enabled("new_ui") {
		t.Error("expected new_ui to be enabled")
	}
}

func TestEnabled_DefaultValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("beta", "true", "Beta features")

	if !m.Enabled("beta") {
		t.Error("expected beta to be enabled by default")
	}
}

func TestEnabled_FalsyValues(t *testing.T) {
	for _, v := range []string{"0", "false", "no", "off", "", "maybe"} {
		m := newTestManagerWithEnv(map[string]string{"FF_FLAG": v})
		m.Register("flag", "false", "test")

		if m.Enabled("flag") {
			t.Errorf("expected %q to be falsy", v)
		}
	}
}

func TestEnabled_TruthyValues(t *testing.T) {
	for _, v := range []string{"1", "true", "TRUE", "True", "yes", "YES", "on", "ON"} {
		m := newTestManagerWithEnv(map[string]string{"FF_FLAG": v})
		m.Register("flag", "false", "test")

		if !m.Enabled("flag") {
			t.Errorf("expected %q to be truthy", v)
		}
	}
}

func TestDisabled(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("off_flag", "false", "Disabled flag")

	if !m.Disabled("off_flag") {
		t.Error("expected off_flag to be disabled")
	}
}

func TestValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MODE": "staging"})
	m.Register("mode", "production", "Deployment mode")

	if got := m.Value("mode"); got != "staging" {
		t.Errorf("expected staging, got %q", got)
	}
}

func TestValue_UsesDefault(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("mode", "production", "Deployment mode")

	if got := m.Value("mode"); got != "production" {
		t.Errorf("expected production, got %q", got)
	}
}

func TestIntValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MAX_RETRIES": "5"})
	m.Register("max_retries", "3", "Max retries")

	if got := m.IntValue("max_retries", 3); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestIntValue_Fallback(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MAX_RETRIES": "not-a-number"})
	m.Register("max_retries", "not-a-number", "Max retries")

	if got := m.IntValue("max_retries", 3); got != 3 {
		t.Errorf("expected fallback 3, got %d", got)
	}
}

func TestFloat64Value(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_RATE": "0.75"})
	m.Register("rate", "0.5", "Rate")

	if got := m.Float64Value("rate", 0.5); got != 0.75 {
		t.Errorf("expected 0.75, got %f", got)
	}
}

func TestFloat64Value_Fallback(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_RATE": "bad"})
	m.Register("rate", "bad", "Rate")

	if got := m.Float64Value("rate", 0.5); got != 0.5 {
		t.Errorf("expected fallback 0.5, got %f", got)
	}
}

func TestListValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_REGIONS": "us-east, eu-west, ap-south"})
	m.Register("regions", "", "Regions")

	list := m.ListValue("regions", ",")
	if len(list) != 3 {
		t.Fatalf("expected 3 items, got %d", len(list))
	}
	if list[0] != "us-east" || list[1] != "eu-west" || list[2] != "ap-south" {
		t.Errorf("unexpected values: %v", list)
	}
}

func TestListValue_Empty(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("regions", "", "Regions")

	list := m.ListValue("regions", ",")
	if list != nil {
		t.Errorf("expected nil for empty value, got %v", list)
	}
}

func TestListValue_SkipsEmptyEntries(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_ITEMS": "a,,b, ,c"})
	m.Register("items", "", "Items")

	list := m.ListValue("items", ",")
	if len(list) != 3 {
		t.Errorf("expected 3 items (skipping empties), got %d: %v", len(list), list)
	}
}

func TestAll(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_A": "1"})
	m.Register("a", "0", "Flag A")
	m.Register("b", "default_b", "Flag B")

	all := m.All()
	if all["a"] != "1" {
		t.Errorf("expected a=1, got %q", all["a"])
	}
	if all["b"] != "default_b" {
		t.Errorf("expected b=default_b, got %q", all["b"])
	}
}

func TestAllFlags(t *testing.T) {
	m := newTestManager()
	m.Register("x", "true", "Flag X")

	flags := m.AllFlags()
	if f, ok := flags["x"]; !ok {
		t.Error("expected flag x")
	} else if f.Description != "Flag X" {
		t.Error("wrong description")
	}
}

func TestUnregisteredFlag_LogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := NewManager(Config{Logger: logger})

	val := m.Value("nonexistent")
	if val != "" {
		t.Errorf("expected empty for unregistered, got %q", val)
	}
	if !bytes.Contains(buf.Bytes(), []byte("unregistered feature flag")) {
		t.Error("expected warning log for unregistered flag")
	}
}

func TestEnabled_UnregisteredFlag(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := NewManager(Config{Logger: logger})

	if m.Enabled("ghost") {
		t.Error("unregistered flag should not be enabled")
	}
}

func TestPrefix_UpperCase(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MY_FLAG": "yes"})
	m.Register("my_flag", "no", "Test flag")

	if !m.Enabled("my_flag") {
		t.Error("expected flag to be enabled via uppercase env var")
	}
}

func BenchmarkEnabled(b *testing.B) {
	m := newTestManagerWithEnv(map[string]string{"FF_BENCH": "true"})
	m.Register("bench", "false", "Bench flag")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Enabled("bench")
	}
}
