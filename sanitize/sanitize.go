// Package sanitize provides input sanitization helpers for filenames and text.
package sanitize

import (
	"path/filepath"
	"strings"
)

// DocFilename removes unsafe characters from a document filename.
func DocFilename(name string) string {
	if name == "" {
		return "unnamed"
	}
	name = filepath.Base(name)
	var clean strings.Builder
	for _, r := range name {
		if r >= 32 && r != 127 {
			clean.WriteRune(r)
		}
	}
	result := clean.String()
	if result == "" {
		result = "unnamed"
	}
	return result
}

// NilIfEmpty returns nil for empty strings, otherwise a pointer to s.
func NilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// IsDuplicateKey checks if a database error is a unique constraint violation (PostgreSQL 23505).
func IsDuplicateKey(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
