package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// macColon matches MAC addresses in colon-separated format (e.g. 01:23:45:67:89:AB).
var macColon = regexp.MustCompile(`^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`)

// macHyphen matches MAC addresses in hyphen-separated format (e.g. 01-23-45-67-89-AB).
var macHyphen = regexp.MustCompile(`^([0-9A-Fa-f]{2}-){5}[0-9A-Fa-f]{2}$`)

// macDot matches MAC addresses in dot-separated format (e.g. 0123.4567.89AB).
var macDot = regexp.MustCompile(`^[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}\.[0-9A-Fa-f]{4}$`)

// MACAddress validates that value is a valid MAC address in any common format
// (colon-separated, hyphen-separated, or dot-separated).
func MACAddress(field, value string) error {
	if macColon.MatchString(value) || macHyphen.MatchString(value) || macDot.MatchString(value) {
		return nil
	}
	return fmt.Errorf("%s must be a valid MAC address", field)
}

// MACAddressColon validates that value is a colon-separated MAC address (e.g. 01:23:45:67:89:AB).
func MACAddressColon(field, value string) error {
	if macColon.MatchString(value) {
		return nil
	}
	return fmt.Errorf("%s must be a colon-separated MAC address", field)
}

// NormalizeMACAddress converts a valid MAC address to uppercase colon-separated format.
// It returns an error if the input is not a valid MAC address.
func NormalizeMACAddress(value string) (string, error) {
	var hex string
	switch {
	case macColon.MatchString(value):
		hex = strings.ReplaceAll(value, ":", "")
	case macHyphen.MatchString(value):
		hex = strings.ReplaceAll(value, "-", "")
	case macDot.MatchString(value):
		hex = strings.ReplaceAll(value, ".", "")
	default:
		return "", fmt.Errorf("invalid MAC address: %s", value)
	}
	hex = strings.ToUpper(hex)
	parts := make([]string, 6)
	for i := range 6 {
		parts[i] = hex[i*2 : i*2+2]
	}
	return strings.Join(parts, ":"), nil
}
