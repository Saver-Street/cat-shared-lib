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

// Float64 reads a float64 environment variable or returns the default.
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

// MustFloat64 reads a required environment variable as a float64.
// Panics if unset, empty, or not a valid floating-point number.
func MustFloat64(key string) float64 {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("config: required environment variable %s is not set", key))
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		panic(fmt.Sprintf("config: environment variable %s=%q is not a valid float", key, v))
	}
	return f
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
