// Package featureflags provides environment-variable-based feature flag
// management. Flags are read from environment variables with a configurable
// prefix and support boolean, percentage-based, and list-of-values evaluation.
package featureflags

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Flag represents a single feature flag with its environment variable name
// and default value.
type Flag struct {
	// EnvVar is the environment variable name (e.g., "FEATURE_NEW_UI").
	EnvVar string
	// DefaultValue is used when the env var is not set.
	DefaultValue string
	// Description documents the flag's purpose.
	Description string
}

// Manager manages a set of feature flags backed by environment variables.
type Manager struct {
	prefix   string
	flags    map[string]Flag
	mu       sync.RWMutex
	logger   *slog.Logger
	resolver func(string) string // injectable for testing
}

// Config configures the feature flag manager.
type Config struct {
	// Prefix is prepended to all flag names when looking up env vars.
	// E.g., prefix "FEATURE_" means flag "new_ui" reads "FEATURE_NEW_UI".
	// Default: "FEATURE_".
	Prefix string
	// Logger for flag evaluation messages. Default: slog.Default().
	Logger *slog.Logger
}

func (c *Config) defaults() {
	if c.Prefix == "" {
		c.Prefix = "FEATURE_"
	}
	if c.Logger == nil {
		c.Logger = slog.Default()
	}
}

// NewManager creates a new feature flag manager.
func NewManager(cfg Config) *Manager {
	cfg.defaults()
	return &Manager{
		prefix:   cfg.Prefix,
		flags:    make(map[string]Flag),
		logger:   cfg.Logger,
		resolver: os.Getenv,
	}
}

// Register adds a flag definition. The envVar will be prefixed with the
// manager's prefix.
func (m *Manager) Register(name, defaultValue, description string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.flags[name] = Flag{
		EnvVar:       m.prefix + strings.ToUpper(name),
		DefaultValue: defaultValue,
		Description:  description,
	}
}

// Enabled returns true if the flag evaluates to a truthy value.
// Truthy values: "1", "true", "yes", "on" (case-insensitive).
func (m *Manager) Enabled(name string) bool {
	val := m.resolve(name)
	return isTruthy(val)
}

// Disabled returns true if the flag is not enabled.
func (m *Manager) Disabled(name string) bool {
	return !m.Enabled(name)
}

// Value returns the raw string value of the flag.
func (m *Manager) Value(name string) string {
	return m.resolve(name)
}

// IntValue returns the flag value as an integer, or the fallback if parsing fails.
func (m *Manager) IntValue(name string, fallback int) int {
	val := m.resolve(name)
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return n
}

// Float64Value returns the flag value as a float64, or the fallback if parsing fails.
func (m *Manager) Float64Value(name string, fallback float64) float64 {
	val := m.resolve(name)
	f, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return fallback
	}
	return f
}

// ListValue returns the flag value split by the given separator.
func (m *Manager) ListValue(name, separator string) []string {
	val := m.resolve(name)
	if val == "" {
		return nil
	}
	parts := strings.Split(val, separator)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// All returns a snapshot of all registered flags and their current values.
func (m *Manager) All() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]string, len(m.flags))
	for name := range m.flags {
		result[name] = m.resolve(name)
	}
	return result
}

// AllFlags returns a copy of all registered flag definitions.
func (m *Manager) AllFlags() map[string]Flag {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make(map[string]Flag, len(m.flags))
	for k, v := range m.flags {
		result[k] = v
	}
	return result
}

func (m *Manager) resolve(name string) string {
	m.mu.RLock()
	flag, ok := m.flags[name]
	m.mu.RUnlock()

	if !ok {
		m.logger.Warn("unregistered feature flag accessed", "flag", name)
		return ""
	}

	val := m.resolver(flag.EnvVar)
	if val == "" {
		return flag.DefaultValue
	}
	return val
}

func isTruthy(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
