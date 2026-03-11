package config

import (
	"fmt"
	"os"
	"strings"
)

// Lookup returns the value of the environment variable and true if it is set
// and non-empty, or ("", false) otherwise.  Unlike [String], it distinguishes
// between an unset variable and one set to "".
func Lookup(key string) (string, bool) {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return "", false
	}
	return v, true
}

// ValidateAll returns an error listing every key in keys that is unset or
// empty.  Use this to enforce that all of a group of related variables are
// provided together (e.g., database host, port, user, password).
func ValidateAll(keys ...string) error {
	var missing []string
	for _, k := range keys {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("config: required keys missing: %s", strings.Join(missing, ", "))
	}
	return nil
}

// ValidateAny returns an error if none of the provided keys is set and
// non-empty.  Use this when at least one of several alternative variables must
// be provided (e.g., API_KEY or API_TOKEN).
func ValidateAny(keys ...string) error {
	for _, k := range keys {
		if os.Getenv(k) != "" {
			return nil
		}
	}
	return fmt.Errorf("config: at least one key required: %s", strings.Join(keys, ", "))
}

// FeatureEnabled returns true if the named environment variable is set to a
// truthy value ("1", "true", "yes", "on") case-insensitively.  Any other value
// (including unset) returns false.
func FeatureEnabled(key string) bool {
	switch strings.ToLower(os.Getenv(key)) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
