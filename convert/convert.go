package convert

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ToInt converts a string to int, returning fallback on failure.
func ToInt(s string, fallback int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

// ToInt64 converts a string to int64, returning fallback on failure.
func ToInt64(s string, fallback int64) int64 {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}

// ToFloat64 converts a string to float64, returning fallback on failure.
func ToFloat64(s string, fallback float64) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fallback
	}
	return v
}

// ToBool converts a string to bool, returning fallback on failure.
// Accepts "true", "1", "yes", "on" as true; "false", "0", "no", "off" as false.
func ToBool(s string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return fallback
	}
}

// ToString converts any basic type to its string representation.
func ToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	case nil:
		return ""
	default:
		return fmt.Sprintf("%v", val)
	}
}

// ToDuration converts a string to time.Duration, returning fallback on failure.
func ToDuration(s string, fallback time.Duration) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return fallback
	}
	return d
}

// ToUint converts a string to uint, returning fallback on failure.
func ToUint(s string, fallback uint) uint {
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return fallback
	}
	return uint(v)
}

// MustInt converts a string to int, panicking on failure.
func MustInt(s string) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		panic("convert: invalid int: " + s)
	}
	return v
}

// MustFloat64 converts a string to float64, panicking on failure.
func MustFloat64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic("convert: invalid float64: " + s)
	}
	return v
}

// Ptr returns a pointer to v. Useful for creating pointers to literals.
func Ptr[T any](v T) *T {
	return &v
}

// Deref returns the value pointed to by p, or the zero value if p is nil.
func Deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// DerefOr returns the value pointed to by p, or fallback if p is nil.
func DerefOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}
