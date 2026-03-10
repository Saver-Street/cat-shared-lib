package sanitize

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
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
	f.Add("23505")
	f.Add("23503")
	f.Add("")
	f.Add("42000")

	f.Fuzz(func(t *testing.T, code string) {
		err := &pgconn.PgError{Code: code}
		result := IsDuplicateKey(err)
		expected := code == "23505"
		if result != expected {
			t.Errorf("IsDuplicateKey(&pgconn.PgError{Code: %q}) = %v, want %v", code, result, expected)
		}
		// Plain errors should never match.
		if IsDuplicateKey(errors.New(code)) {
			t.Errorf("IsDuplicateKey(errors.New(%q)) should be false", code)
		}
	})
}

func FuzzTruncateFilename(f *testing.F) {
	f.Add("report.pdf", 20)
	f.Add("averylongfilename.txt", 10)
	f.Add("", 10)
	f.Add("file.txt", 0)
	f.Add("file.txt", -1)
	f.Add("x.verylongext", 3)
	f.Add("résumé.pdf", 7)
	f.Add("日本語.txt", 5)
	f.Add("noext", 3)
	f.Add(".gitignore", 5)

	f.Fuzz(func(t *testing.T, name string, maxLen int) {
		result := TruncateFilename(name, maxLen)
		runes := []rune(result)
		if maxLen <= 0 || name == "" {
			if result != "" {
				t.Errorf("TruncateFilename(%q, %d) = %q, want empty", name, maxLen, result)
			}
			return
		}
		nameRunes := []rune(name)
		if len(nameRunes) <= maxLen {
			if result != name {
				t.Errorf("TruncateFilename(%q, %d) = %q, want unchanged", name, maxLen, result)
			}
			return
		}
		// When truncated, length should be at most maxLen (unless ext is longer)
		ext := []rune(filepath.Ext(name))
		if len(ext) < maxLen && len(runes) > maxLen {
			t.Errorf("TruncateFilename(%q, %d) = %q (len %d), exceeds maxLen", name, maxLen, result, len(runes))
		}
	})
}

func FuzzMaxLength(f *testing.F) {
	f.Add("hello world", 5)
	f.Add("hello", 10)
	f.Add("", 5)
	f.Add("hello", 0)
	f.Add("hello", -1)
	f.Add("héllo", 3)
	f.Add("日本語テスト", 3)

	f.Fuzz(func(t *testing.T, s string, maxLen int) {
		result := MaxLength(s, maxLen)
		runes := []rune(result)
		if maxLen <= 0 {
			if result != "" {
				t.Errorf("MaxLength(%q, %d) = %q, want empty", s, maxLen, result)
			}
			return
		}
		if len(runes) > maxLen {
			t.Errorf("MaxLength(%q, %d) = %q (len %d), exceeds maxLen", s, maxLen, result, len(runes))
		}
		sRunes := []rune(s)
		if len(sRunes) <= maxLen && result != s {
			t.Errorf("MaxLength(%q, %d) = %q, want unchanged", s, maxLen, result)
		}
	})
}
