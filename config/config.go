// Package config provides helpers for loading configuration from
// environment variables with defaults and validation.
package config

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// String reads an environment variable or returns the default value.
func String(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// StringRequired reads an environment variable and returns an error if unset or empty.
func StringRequired(key string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		return "", fmt.Errorf("config: required environment variable %s is not set", key)
	}
	return v, nil
}

// Int reads an integer environment variable or returns the default.
func Int(key string, defaultVal int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return defaultVal
	}
	return n
}

// Int64 reads an int64 environment variable or returns the default.
func Int64(key string, defaultVal int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return defaultVal
	}
	return n
}

// Bool reads a boolean environment variable or returns the default.
// Truthy values: "true", "1", "yes" (case-insensitive).
func Bool(key string, defaultVal bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	return defaultVal
}

// Duration reads a duration string from an environment variable or returns the default.
func Duration(key string, defaultVal time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return defaultVal
	}
	return d
}

// StringSlice reads a comma-separated environment variable into a string slice.
// Returns the default if the variable is unset or empty.
func StringSlice(key string, defaultVal []string) []string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}

// MustString reads an environment variable and panics if it is unset or empty.
func MustString(key string) string {
	v, err := StringRequired(key)
	if err != nil {
		panic(err)
	}
	return v
}

// Float64 reads an environment variable as a float64 or returns the default value.
func Float64(key string, defaultVal float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return defaultVal
	}
	return f
}

// MustFloat64 reads a float64 environment variable and panics if unset or unparseable.
func MustFloat64(key string) float64 {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(fmt.Sprintf("config: %s is not a valid float64: %s", key, v))
	}
	return f
}

// FilePath reads an environment variable and validates that the file exists.
// Returns the default if the variable is unset. Returns an error if the file
// does not exist or is a directory.
func FilePath(key, defaultVal string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		v = defaultVal
	}
	if v == "" {
		return "", nil
	}
	info, err := os.Stat(v)
	if err != nil {
		return "", fmt.Errorf("config: %s: file %q does not exist", key, v)
	}
	if info.IsDir() {
		return "", fmt.Errorf("config: %s: %q is a directory, not a file", key, v)
	}
	return v, nil
}

// DirPath reads an environment variable and validates that the directory exists.
// Returns the default if the variable is unset. Returns an error if the path
// does not exist or is not a directory.
func DirPath(key, defaultVal string) (string, error) {
	v := os.Getenv(key)
	if v == "" {
		v = defaultVal
	}
	if v == "" {
		return "", nil
	}
	info, err := os.Stat(v)
	if err != nil {
		return "", fmt.Errorf("config: %s: directory %q does not exist", key, v)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("config: %s: %q is not a directory", key, v)
	}
	return v, nil
}

// Prefix returns a scoped set of config helpers that prepend a prefix to every
// environment variable key. This is useful for namespacing configuration, e.g.
// Prefix("DB_") turns calls like Get("HOST") into lookups of "DB_HOST".
type Prefix string

// String reads prefix+key or returns the default value.
func (p Prefix) String(key, defaultVal string) string {
	return String(string(p)+key, defaultVal)
}

// Int reads prefix+key as an int or returns the default value.
func (p Prefix) Int(key string, defaultVal int) int {
	return Int(string(p)+key, defaultVal)
}

// Bool reads prefix+key as a bool or returns the default value.
func (p Prefix) Bool(key string, defaultVal bool) bool {
	return Bool(string(p)+key, defaultVal)
}

// Duration reads prefix+key as a time.Duration or returns the default value.
func (p Prefix) Duration(key string, defaultVal time.Duration) time.Duration {
	return Duration(string(p)+key, defaultVal)
}

// MustString reads prefix+key and panics if unset.
func (p Prefix) MustString(key string) string {
	return MustString(string(p) + key)
}

// Summary returns a sorted key-value map of the given environment variable
// names and their current values. Keys containing "SECRET", "PASSWORD",
// "TOKEN", or "KEY" (case-insensitive) have their values masked. This is
// intended for logging configuration at startup.
func Summary(keys ...string) map[string]string {
	m := make(map[string]string, len(keys))
	sort.Strings(keys)
	for _, k := range keys {
		v := os.Getenv(k)
		if v == "" {
			m[k] = "(unset)"
			continue
		}
		upper := strings.ToUpper(k)
		if strings.Contains(upper, "SECRET") ||
			strings.Contains(upper, "PASSWORD") ||
			strings.Contains(upper, "TOKEN") ||
			strings.Contains(upper, "KEY") {
			m[k] = "****"
		} else {
			m[k] = v
		}
	}
	return m
}

// MustInt reads a required environment variable as an int.
// Panics if unset, empty, or not a valid integer.
func MustInt(key string) int {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("config: environment variable %s=%q is not a valid integer", key, v))
	}
	return n
}

// MustInt64 reads a required int64 environment variable.
// Panics if unset, empty, or not a valid int64.
func MustInt64(key string) int64 {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("config: environment variable %s=%q is not a valid int64", key, v))
	}
	return n
}
