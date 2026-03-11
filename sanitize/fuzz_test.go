package sanitize

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
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

func FuzzStripHTML(f *testing.F) {
	f.Add("<script>alert('xss')</script>")
	f.Add("<b>bold</b> and <i>italic</i>")
	f.Add("")
	f.Add("no tags here")
	f.Add("<img src=x onerror=alert(1)>")
	f.Add("<<nested>>text<</nested>>")
	f.Add("<div\x00onmouseover=alert(1)>hi</div>")
	f.Add("text < 5 && x > 3")
	f.Add("<a href=\"javascript:alert(1)\">click</a>")
	f.Add("<svg/onload=alert(1)>")
	f.Add("日本語<b>テスト</b>")
	f.Add(">\x00<script\x00>alert\x00</script\x00>")

	f.Fuzz(func(t *testing.T, s string) {
		result := StripHTML(s)
		// Result must not contain complete HTML tags.
		inTag := false
		for _, r := range result {
			if r == '<' {
				inTag = true
			} else if r == '>' && inTag {
				t.Errorf("StripHTML(%q) still contains tag-like sequence in %q",
					truncate(s, 200), truncate(result, 200))
				return
			}
		}
		// Result rune count should never exceed input rune count.
		if len([]rune(result)) > len([]rune(s)) {
			t.Errorf("StripHTML rune count grew: %d > %d", len([]rune(result)), len([]rune(s)))
		}
	})
}

func FuzzSanitizeEmail(f *testing.F) {
	f.Add("USER@EXAMPLE.COM")
	f.Add("  alice@example.com  ")
	f.Add("")
	f.Add("\t\n user@host.co \r\n")
	f.Add("O'Brien@example.com")
	f.Add("user+tag@example.com")
	f.Add("用户@例え.jp")
	f.Add("a]b[c@d<e>f")
	f.Add("\x00admin@evil.com")
	f.Add("ADMIN@EXAMPLE.COM\x00regular@example.com")

	f.Fuzz(func(t *testing.T, email string) {
		result := SanitizeEmail(email)
		// Must be fully lowercased (same as applying ToLower again).
		lower := strings.ToLower(result)
		if result != lower {
			t.Errorf("SanitizeEmail(%q) not fully lowercase: %q vs %q", email, result, lower)
		}
		// Must not have leading/trailing whitespace.
		if result != strings.TrimSpace(result) {
			t.Errorf("SanitizeEmail(%q) has untrimmed whitespace: %q", email, result)
		}
		// Idempotent: sanitizing twice yields the same result.
		if double := SanitizeEmail(result); double != result {
			t.Errorf("SanitizeEmail not idempotent: %q -> %q -> %q", email, result, double)
		}
	})
}

func FuzzTrimAndNilIfEmpty(f *testing.F) {
	f.Add("")
	f.Add("   ")
	f.Add("\t\n\r")
	f.Add("hello")
	f.Add("  hello  ")
	f.Add("\x00")
	f.Add("日本語")
	f.Add("  \t  value  \n  ")

	f.Fuzz(func(t *testing.T, s string) {
		result := TrimAndNilIfEmpty(s)
		trimmed := strings.TrimSpace(s)
		if trimmed == "" {
			if result != nil {
				t.Errorf("TrimAndNilIfEmpty(%q) should return nil for whitespace-only", s)
			}
			return
		}
		if result == nil {
			t.Fatalf("TrimAndNilIfEmpty(%q) returned nil for non-empty trimmed %q", s, trimmed)
		}
		if *result != trimmed {
			t.Errorf("TrimAndNilIfEmpty(%q) = %q, want %q", s, *result, trimmed)
		}
	})
}

func FuzzIsDatabaseError(f *testing.F) {
	f.Add("23505", "23505")
	f.Add("23503", "23505")
	f.Add("", "")
	f.Add("42000", "23503")
	f.Add("23505", "")
	f.Add("\x00", "23505")

	f.Fuzz(func(t *testing.T, errCode, checkCode string) {
		pgErr := &pgconn.PgError{Code: errCode}
		result := IsDatabaseError(pgErr, checkCode)
		expected := errCode == checkCode
		if result != expected {
			t.Errorf("IsDatabaseError(code=%q, check=%q) = %v, want %v",
				errCode, checkCode, result, expected)
		}
		// Plain errors must never match.
		if IsDatabaseError(errors.New("some error"), checkCode) {
			t.Errorf("IsDatabaseError(plain error, %q) should be false", checkCode)
		}
		if IsDatabaseError(nil, checkCode) {
			t.Errorf("IsDatabaseError(nil, %q) should be false", checkCode)
		}
	})
}

func FuzzDeref(f *testing.F) {
	f.Add("value", "default", true)
	f.Add("", "fallback", true)
	f.Add("", "fallback", false)
	f.Add("hello", "", false)

	f.Fuzz(func(t *testing.T, val, def string, usePtr bool) {
		if usePtr {
			result := Deref(&val, def)
			if result != val {
				t.Errorf("Deref(&%q, %q) = %q, want %q", val, def, result, val)
			}
		} else {
			result := Deref[string](nil, def)
			if result != def {
				t.Errorf("Deref(nil, %q) = %q, want %q", def, result, def)
			}
		}
	})
}

func FuzzDocFilenameConcurrent(f *testing.F) {
	f.Add("report.pdf", 8)
	f.Add("../etc/passwd", 4)
	f.Add("", 16)
	f.Add("\x00\x01\x02", 2)
	f.Add("résumé_日本語.pdf", 6)

	f.Fuzz(func(t *testing.T, name string, goroutines int) {
		if goroutines <= 0 {
			goroutines = 1
		}
		if goroutines > 32 {
			goroutines = 32
		}
		var wg sync.WaitGroup
		results := make([]string, goroutines)
		wg.Add(goroutines)
		for i := range goroutines {
			go func(idx int) {
				defer wg.Done()
				results[idx] = DocFilename(name)
			}(i)
		}
		wg.Wait()
		// All goroutines must produce the same result.
		for i := 1; i < goroutines; i++ {
			if results[i] != results[0] {
				t.Errorf("concurrent DocFilename(%q): goroutine %d got %q, goroutine 0 got %q",
					name, i, results[i], results[0])
			}
		}
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return fmt.Sprintf("%s...(len=%d)", s[:n], len(s))
}
