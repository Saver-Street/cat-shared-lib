package sanitize

import (
	"strings"
	"unicode/utf8"
)

// TruncateWords returns the first n words of s, appending "..." if truncated.
// Words are split on whitespace.
func TruncateWords(s string, n int) string {
	if n <= 0 {
		return ""
	}
	words := strings.Fields(s)
	if len(words) <= n {
		return s
	}
	return strings.Join(words[:n], " ") + "..."
}

// PadLeft pads s on the left with pad until s reaches totalWidth runes.
// If s is already at least totalWidth, it is returned unchanged.
func PadLeft(s string, totalWidth int, pad rune) string {
	n := utf8.RuneCountInString(s)
	if n >= totalWidth {
		return s
	}
	return strings.Repeat(string(pad), totalWidth-n) + s
}

// PadRight pads s on the right with pad until s reaches totalWidth runes.
// If s is already at least totalWidth, it is returned unchanged.
func PadRight(s string, totalWidth int, pad rune) string {
	n := utf8.RuneCountInString(s)
	if n >= totalWidth {
		return s
	}
	return s + strings.Repeat(string(pad), totalWidth-n)
}

// ReverseString reverses s by runes, correctly handling multi-byte characters.
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Excerpt extracts a substring of maxLen runes centered around the first
// occurrence of phrase. It prepends "..." if the excerpt doesn't start at
// the beginning and appends "..." if it doesn't end at the end.
// If phrase is not found, it returns the first maxLen runes.
func Excerpt(s, phrase string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	phraseIdx := strings.Index(strings.ToLower(s), strings.ToLower(phrase))
	if phraseIdx < 0 {
		result := string(runes[:maxLen])
		return result + "..."
	}

	// Convert byte index to rune index
	runeIdx := utf8.RuneCountInString(s[:phraseIdx])

	start := runeIdx - maxLen/2
	if start < 0 {
		start = 0
	}
	end := start + maxLen
	if end > len(runes) {
		end = len(runes)
		start = end - maxLen
	}

	result := string(runes[start:end])
	if start > 0 {
		result = "..." + result
	}
	if end < len(runes) {
		result = result + "..."
	}
	return result
}

// WordWrap wraps s at lineWidth characters, breaking on whitespace.
// If a single word exceeds lineWidth, it is placed on its own line.
func WordWrap(s string, lineWidth int) string {
	if lineWidth <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	lineLen := 0
	for i, word := range words {
		wLen := utf8.RuneCountInString(word)
		if i > 0 && lineLen+1+wLen > lineWidth {
			b.WriteByte('\n')
			lineLen = 0
		} else if i > 0 {
			b.WriteByte(' ')
			lineLen++
		}
		b.WriteString(word)
		lineLen += wLen
	}
	return b.String()
}
