package validation

import (
	"fmt"
	"strings"
)

// FQDN validates that value is a fully qualified domain name. An FQDN must
// contain at least two labels separated by dots, with each label following
// RFC 1123 hostname rules. A trailing dot is permitted but not required.
func FQDN(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return fmt.Errorf("%s must be a valid FQDN", field)
	}
	// Allow trailing dot
	v = strings.TrimSuffix(v, ".")
	if v == "" {
		return fmt.Errorf("%s must be a valid FQDN", field)
	}
	if len(v) > 253 {
		return fmt.Errorf("%s must be a valid FQDN", field)
	}

	labels := strings.Split(v, ".")
	if len(labels) < 2 {
		return fmt.Errorf("%s must be a valid FQDN", field)
	}

	for _, label := range labels {
		if len(label) == 0 || len(label) > 63 {
			return fmt.Errorf("%s must be a valid FQDN", field)
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return fmt.Errorf("%s must be a valid FQDN", field)
		}
		for _, c := range label {
			if !isAlphaNumericHyphen(c) {
				return fmt.Errorf("%s must be a valid FQDN", field)
			}
		}
	}

	// TLD must not be all-numeric
	tld := labels[len(labels)-1]
	allDigits := true
	for _, c := range tld {
		if c < '0' || c > '9' {
			allDigits = false
			break
		}
	}
	if allDigits {
		return fmt.Errorf("%s must be a valid FQDN", field)
	}

	return nil
}

func isAlphaNumericHyphen(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-'
}

// DataURI validates that value is a well-formed data URI (RFC 2397).
// It checks the data: prefix, media type, and optional base64 indicator.
func DataURI(field, value string) error {
	if !strings.HasPrefix(value, "data:") {
		return fmt.Errorf("%s must be a valid data URI", field)
	}
	rest := value[5:]
	commaIdx := strings.IndexByte(rest, ',')
	if commaIdx < 0 {
		return fmt.Errorf("%s must be a valid data URI", field)
	}
	meta := rest[:commaIdx]

	// Validate media type if present
	if meta != "" && meta != "base64" {
		// Check for ;base64 suffix
		base64Suffix := strings.HasSuffix(meta, ";base64")
		if base64Suffix {
			meta = meta[:len(meta)-7]
		}
		// Media type should contain a /
		if meta != "" && !strings.Contains(meta, "/") {
			return fmt.Errorf("%s must be a valid data URI", field)
		}
	}

	return nil
}
