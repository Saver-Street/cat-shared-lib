package stringutil

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ToKebabCase converts camelCase/PascalCase/snake_case to kebab-case.
func ToKebabCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)
	for i, r := range s {
		if r == '_' || r == ' ' {
			if b.Len() > 0 {
				b.WriteByte('-')
			}
			continue
		}
		if unicode.IsUpper(r) {
			if i > 0 {
				prev, _ := utf8.DecodeLastRuneInString(s[:i])
				if prev != '_' && prev != ' ' && prev != '-' && !unicode.IsUpper(prev) {
					b.WriteByte('-')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// ToPascalCase converts snake_case/kebab-case/camelCase to PascalCase.
func ToPascalCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	capNext := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			capNext = true
			continue
		}
		if capNext {
			b.WriteRune(unicode.ToUpper(r))
			capNext = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// PadLeft pads s on the left with the pad rune to reach length n.
// If s is already at least n runes long, it is returned unchanged.
func PadLeft(s string, n int, pad rune) string {
	rl := utf8.RuneCountInString(s)
	if rl >= n {
		return s
	}
	var b strings.Builder
	b.Grow(n * utf8.UTFMax)
	for range n - rl {
		b.WriteRune(pad)
	}
	b.WriteString(s)
	return b.String()
}

// PadRight pads s on the right with the pad rune to reach length n.
// If s is already at least n runes long, it is returned unchanged.
func PadRight(s string, n int, pad rune) string {
	rl := utf8.RuneCountInString(s)
	if rl >= n {
		return s
	}
	var b strings.Builder
	b.Grow(n * utf8.UTFMax)
	b.WriteString(s)
	for range n - rl {
		b.WriteRune(pad)
	}
	return b.String()
}

// IsBlank returns true if s is empty or contains only whitespace.
func IsBlank(s string) bool {
	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// Reverse returns s with its runes in reverse order.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// WordWrap wraps s at the given line width, breaking at word boundaries.
// Words longer than width are placed on their own line without breaking.
func WordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s) + len(words))
	lineLen := 0
	for i, w := range words {
		wl := utf8.RuneCountInString(w)
		if i == 0 {
			b.WriteString(w)
			lineLen = wl
			continue
		}
		if lineLen+1+wl > width {
			b.WriteByte('\n')
			b.WriteString(w)
			lineLen = wl
		} else {
			b.WriteByte(' ')
			b.WriteString(w)
			lineLen += 1 + wl
		}
	}
	return b.String()
}

// CountWords returns the number of whitespace-separated words in s.
func CountWords(s string) int {
	return len(strings.Fields(s))
}
