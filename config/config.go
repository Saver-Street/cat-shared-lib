// Package config provides helpers for loading configuration from
// environment variables with defaults and validation.
package config

import (
	"fmt"
	"os"
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

// MustBool reads a required boolean environment variable.
// Panics if unset, empty, or not a recognized boolean value.
// Truthy: "true", "1", "yes". Falsy: "false", "0", "no" (case-insensitive).
func MustBool(key string) bool {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	panic(fmt.Sprintf("config: environment variable %s=%q is not a valid boolean", key, v))
}

// MustDuration reads a required environment variable as a time.Duration.
// Panics if unset, empty, or not a valid duration string.
func MustDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("config: environment variable %s=%q is not a valid duration", key, v))
	}
	return d
}

// Validate checks that all required keys are set and non-empty.
// Returns a combined error listing all missing keys.
func Validate(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("config: missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
}

// StringMap reads a comma-separated list of key=value pairs from an
// environment variable. Returns the default if the variable is unset or empty.
// Pairs missing an "=" sign are silently skipped. Both keys and values are trimmed.
//
// Example: LABELS="env=prod, region=us-east" → map[string]string{"env":"prod", "region":"us-east"}
func StringMap(key string, defaultVal map[string]string) map[string]string {
	v := os.Getenv(key)
	if v == "" {
		return defaultVal
	}
	result := make(map[string]string)
	for _, pair := range strings.Split(v, ",") {
		k, val, ok := strings.Cut(strings.TrimSpace(pair), "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		val = strings.TrimSpace(val)
		if k != "" {
			result[k] = val
		}
	}
	if len(result) == 0 {
		return defaultVal
	}
	return result
}
