// Package config provides helpers for loading configuration from
// environment variables with defaults and validation.
package config

import (
	"fmt"
	"net"
	"net/url"
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

// URL reads a URL from the named environment variable, validates that it is
// a well-formed absolute URL with an http or https scheme, and returns it.
// If the variable is unset or empty, defaultVal is returned.
func URL(key string, defaultVal string) (string, error) {
v := strings.TrimSpace(os.Getenv(key))
if v == "" {
return defaultVal, nil
}
u, err := url.Parse(v)
if err != nil {
return "", fmt.Errorf("config: %s: invalid URL: %w", key, err)
}
if u.Scheme != "http" && u.Scheme != "https" {
return "", fmt.Errorf("config: %s: URL must use http or https scheme", key)
}
if u.Host == "" {
return "", fmt.Errorf("config: %s: URL must have a host", key)
}
return v, nil
}

// Port reads a TCP port number from the named environment variable. It
// validates that the value is between 1 and 65535. If the variable is
// unset or empty, defaultVal is returned.
func Port(key string, defaultVal int) (int, error) {
v := strings.TrimSpace(os.Getenv(key))
if v == "" {
return defaultVal, nil
}
p, err := strconv.Atoi(v)
if err != nil {
return 0, fmt.Errorf("config: %s: invalid port number: %w", key, err)
}
if p < 1 || p > 65535 {
return 0, fmt.Errorf("config: %s: port must be between 1 and 65535, got %d", key, p)
}
return p, nil
}

// Addr reads a host:port address from the named environment variable. It
// validates that the value has a valid host and port component using
// net.SplitHostPort. If the variable is unset or empty, defaultVal is returned.
func Addr(key, defaultVal string) (string, error) {
v := strings.TrimSpace(os.Getenv(key))
if v == "" {
return defaultVal, nil
}
host, portStr, err := net.SplitHostPort(v)
if err != nil {
return "", fmt.Errorf("config: %s: invalid address: %w", key, err)
}
p, err := strconv.Atoi(portStr)
if err != nil || p < 1 || p > 65535 {
return "", fmt.Errorf("config: %s: port must be between 1 and 65535", key)
}
return net.JoinHostPort(host, portStr), nil
}

// StringSliceRequired reads a comma-separated environment variable into a
// string slice and returns an error if the variable is unset, empty, or
// contains no non-empty elements after trimming.
func StringSliceRequired(key string) ([]string, error) {
v := os.Getenv(key)
if v == "" {
return nil, fmt.Errorf("config: %s is required", key)
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
return nil, fmt.Errorf("config: %s is required", key)
}
return result, nil
}

// Bytes reads a byte-size value from the named environment variable.
// Supported suffixes (case-insensitive): B, KB, MB, GB, TB.
// Plain integers are treated as bytes. If the variable is unset or empty,
// defaultVal is returned.
func Bytes(key string, defaultVal int64) (int64, error) {
v := strings.TrimSpace(os.Getenv(key))
if v == "" {
return defaultVal, nil
}
upper := strings.ToUpper(v)
multiplier := int64(1)
numStr := v
switch {
case strings.HasSuffix(upper, "TB"):
multiplier = 1 << 40
numStr = strings.TrimSpace(v[:len(v)-2])
case strings.HasSuffix(upper, "GB"):
multiplier = 1 << 30
numStr = strings.TrimSpace(v[:len(v)-2])
case strings.HasSuffix(upper, "MB"):
multiplier = 1 << 20
numStr = strings.TrimSpace(v[:len(v)-2])
case strings.HasSuffix(upper, "KB"):
multiplier = 1 << 10
numStr = strings.TrimSpace(v[:len(v)-2])
case strings.HasSuffix(upper, "B"):
numStr = strings.TrimSpace(v[:len(v)-1])
}
n, err := strconv.ParseInt(numStr, 10, 64)
if err != nil {
return 0, fmt.Errorf("config: %s: invalid byte size %q: %w", key, v, err)
}
return n * multiplier, nil
}

// Enum reads key from the environment and validates that its value is one
// of the allowed choices. If the variable is unset or empty, defaultVal is
// returned. Comparison is case-sensitive.
func Enum(key, defaultVal string, allowed []string) (string, error) {
v := strings.TrimSpace(os.Getenv(key))
if v == "" {
return defaultVal, nil
}
for _, a := range allowed {
if v == a {
return v, nil
}
}
return "", fmt.Errorf("config: %s: value %q is not one of %v", key, v, allowed)
}

// MustEnum is like Enum but panics if the value is not one of the allowed
// choices. Intended for use during application startup.
func MustEnum(key string, allowed []string) string {
v, err := Enum(key, "", allowed)
if err != nil {
panic(err)
}
if v == "" {
panic(fmt.Sprintf("config: %s is required", key))
}
return v
}

// MustURL is like URL but panics if the variable is unset or the value is
// not a valid HTTP/HTTPS URL. Intended for use during application startup.
func MustURL(key string) string {
v, err := URL(key, "")
if err != nil {
panic(err)
}
if v == "" {
panic(fmt.Sprintf("config: %s is required", key))
}
return v
}

// MustPort is like Port but panics if the variable is unset or the value is
// not a valid port (1–65535). Intended for use during application startup.
func MustPort(key string) int {
v, err := Port(key, 0)
if err != nil {
panic(err)
}
if v == 0 {
panic(fmt.Sprintf("config: %s is required", key))
}
return v
}

// MustAddr is like Addr but panics if the variable is unset or the value is
// not a valid host:port address. Intended for use during application startup.
func MustAddr(key string) string {
v, err := Addr(key, "")
if err != nil {
panic(err)
}
if v == "" {
panic(fmt.Sprintf("config: %s is required", key))
}
return v
}
