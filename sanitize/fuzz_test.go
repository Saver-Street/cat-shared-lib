package sanitize

import (
	"errors"
	"strings"
	"testing"
)

func FuzzDocFilename(f *testing.F) {
	f.Add("report.pdf")
	f.Add("")
	f.Add("../../../etc/passwd")
	f.Add("file\x00name\x01.txt")
	f.Add("\x01\x02\x03")
	f.Add("résumé_日本語.pdf")
	f.Add("📄document.pdf")
	f.Add(".gitignore")
	f.Add("file\x7fname.txt")
	f.Add("   ")

	f.Fuzz(func(t *testing.T, name string) {
		result := DocFilename(name)
		// Should never return empty string
		if result == "" {
			t.Errorf("DocFilename(%q) returned empty string", name)
		}
		// Should never contain control characters (< 32 or 127)
		for _, r := range result {
			if r < 32 || r == 127 {
				t.Errorf("DocFilename(%q) contains control character %d", name, r)
			}
		}
	})
}

func FuzzNilIfEmpty(f *testing.F) {
	f.Add("")
	f.Add("hello")
	f.Add(" ")
	f.Add("\t")

	f.Fuzz(func(t *testing.T, s string) {
		result := NilIfEmpty(s)
		if s == "" {
			if result != nil {
				t.Error("NilIfEmpty(\"\") should return nil")
			}
		} else {
			if result == nil {
				t.Fatalf("NilIfEmpty(%q) should not return nil", s)
			}
			if *result != s {
				t.Errorf("NilIfEmpty(%q) = %q, want %q", s, *result, s)
			}
		}
	})
}

func FuzzIsDuplicateKey(f *testing.F) {
	f.Add("duplicate key value violates unique constraint (SQLSTATE 23505)")
	f.Add("ERROR: 23505")
	f.Add("some other error")
	f.Add("")
	f.Add("2350")
	f.Add("23505")

	f.Fuzz(func(t *testing.T, s string) {
		err := errors.New(s)
		result := IsDuplicateKey(err)
		expected := strings.Contains(s, "23505")
		if result != expected {
			t.Errorf("IsDuplicateKey(errors.New(%q)) = %v, want %v", s, result, expected)
		}
	})
}
