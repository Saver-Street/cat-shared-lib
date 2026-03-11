package featureflags

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
	testkit.AssertEqual(t, m.prefix, "FEATURE_")
	testkit.AssertNotNil(t, m.logger)
}

func TestRegister_And_Enabled(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_NEW_UI": "true"})
	m.Register("new_ui", "false", "Enable new UI")

	testkit.AssertTrue(t, m.Enabled("new_ui"))
}

func TestEnabled_DefaultValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("beta", "true", "Beta features")

	testkit.AssertTrue(t, m.Enabled("beta"))
}

func TestEnabled_FalsyValues(t *testing.T) {
	for _, v := range []string{"0", "false", "no", "off", "", "maybe"} {
		m := newTestManagerWithEnv(map[string]string{"FF_FLAG": v})
		m.Register("flag", "false", "test")

		testkit.AssertFalse(t, m.Enabled("flag"))
	}
}

func TestEnabled_TruthyValues(t *testing.T) {
	for _, v := range []string{"1", "true", "TRUE", "True", "yes", "YES", "on", "ON"} {
		m := newTestManagerWithEnv(map[string]string{"FF_FLAG": v})
		m.Register("flag", "false", "test")

		testkit.AssertTrue(t, m.Enabled("flag"))
	}
}

func TestDisabled(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("off_flag", "false", "Disabled flag")

	testkit.AssertTrue(t, m.Disabled("off_flag"))
}

func TestValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MODE": "staging"})
	m.Register("mode", "production", "Deployment mode")

	testkit.AssertEqual(t, m.Value("mode"), "staging")
}

func TestValue_UsesDefault(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("mode", "production", "Deployment mode")

	testkit.AssertEqual(t, m.Value("mode"), "production")
}

func TestIntValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MAX_RETRIES": "5"})
	m.Register("max_retries", "3", "Max retries")

	testkit.AssertEqual(t, m.IntValue("max_retries", 3), 5)
}

func TestIntValue_Fallback(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MAX_RETRIES": "not-a-number"})
	m.Register("max_retries", "not-a-number", "Max retries")

	testkit.AssertEqual(t, m.IntValue("max_retries", 3), 3)
}

func TestFloat64Value(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_RATE": "0.75"})
	m.Register("rate", "0.5", "Rate")

	testkit.AssertEqual(t, m.Float64Value("rate", 0.5), 0.75)
}

func TestFloat64Value_Fallback(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_RATE": "bad"})
	m.Register("rate", "bad", "Rate")

	testkit.AssertEqual(t, m.Float64Value("rate", 0.5), 0.5)
}

func TestListValue(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_REGIONS": "us-east, eu-west, ap-south"})
	m.Register("regions", "", "Regions")

	list := m.ListValue("regions", ",")
	testkit.RequireLen(t, list, 3)
	testkit.AssertEqual(t, list[0], "us-east")
	testkit.AssertEqual(t, list[1], "eu-west")
	testkit.AssertEqual(t, list[2], "ap-south")
}

func TestListValue_Empty(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{})
	m.Register("regions", "", "Regions")

	list := m.ListValue("regions", ",")
	testkit.AssertNil(t, list)
}

func TestListValue_SkipsEmptyEntries(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_ITEMS": "a,,b, ,c"})
	m.Register("items", "", "Items")

	list := m.ListValue("items", ",")
	testkit.AssertLen(t, list, 3)
}

func TestAll(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_A": "1"})
	m.Register("a", "0", "Flag A")
	m.Register("b", "default_b", "Flag B")

	all := m.All()
	testkit.AssertEqual(t, all["a"], "1")
	testkit.AssertEqual(t, all["b"], "default_b")
}

func TestAllFlags(t *testing.T) {
	m := newTestManager()
	m.Register("x", "true", "Flag X")

	flags := m.AllFlags()
	if f, ok := flags["x"]; !ok {
		t.Error("expected flag x")
	} else {
		testkit.AssertEqual(t, f.Description, "Flag X")
	}
}

func TestUnregisteredFlag_LogsWarning(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := NewManager(Config{Logger: logger})

	val := m.Value("nonexistent")
	testkit.AssertEqual(t, val, "")
	testkit.AssertContains(t, buf.String(), "unregistered feature flag")
}

func TestEnabled_UnregisteredFlag(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	m := NewManager(Config{Logger: logger})

	testkit.AssertFalse(t, m.Enabled("ghost"))
}

func TestPrefix_UpperCase(t *testing.T) {
	m := newTestManagerWithEnv(map[string]string{"FF_MY_FLAG": "yes"})
	m.Register("my_flag", "no", "Test flag")

	testkit.AssertTrue(t, m.Enabled("my_flag"))
}

func BenchmarkEnabled(b *testing.B) {
	m := newTestManagerWithEnv(map[string]string{"FF_BENCH": "true"})
	m.Register("bench", "false", "Bench flag")
	b.ResetTimer()
	for b.Loop() {
		m.Enabled("bench")
	}
}
