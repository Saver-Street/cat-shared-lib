package validation

import (
	"fmt"
	"path/filepath"
	"strings"
)

// FileExtension validates that value has one of the allowed file extensions.
// Extensions should include the leading dot (e.g. ".jpg", ".png").
func FileExtension(field, value string, allowed []string) error {
	ext := strings.ToLower(filepath.Ext(value))
	for _, a := range allowed {
		if ext == strings.ToLower(a) {
			return nil
		}
	}
	return fmt.Errorf("%s must have one of the following extensions: %s", field, strings.Join(allowed, ", "))
}

// MIMEType validates that value is one of the allowed MIME types.
// Comparison is case-insensitive.
func MIMEType(field, value string, allowed []string) error {
	v := strings.ToLower(strings.TrimSpace(value))
	for _, a := range allowed {
		if v == strings.ToLower(a) {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of the following MIME types: %s", field, strings.Join(allowed, ", "))
}

// FileSize validates that size (in bytes) does not exceed maxBytes.
func FileSize(field string, size, maxBytes int64) error {
	if size > maxBytes {
		return fmt.Errorf("%s exceeds maximum file size of %d bytes", field, maxBytes)
	}
	if size < 0 {
		return fmt.Errorf("%s has invalid file size", field)
	}
	return nil
}

// SafeFilename validates that value is a safe filename — no path separators,
// no directory traversal, and no control characters.
func SafeFilename(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s must not be empty", field)
	}
	if value == "." || value == ".." {
		return fmt.Errorf("%s must be a safe filename", field)
	}
	if strings.ContainsAny(value, "/\\") {
		return fmt.Errorf("%s must not contain path separators", field)
	}
	if strings.Contains(value, "..") {
		return fmt.Errorf("%s must not contain directory traversal", field)
	}
	for _, c := range value {
		if c < 32 {
			return fmt.Errorf("%s must not contain control characters", field)
		}
	}
	return nil
}
