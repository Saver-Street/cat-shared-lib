package sanitize

import (
	"regexp"
	"strings"
)

// Markdown-stripping patterns.
var (
	mdHeading    = regexp.MustCompile(`(?m)^#{1,6}\s+`)
	mdBold3Star  = regexp.MustCompile(`\*{3}(.+?)\*{3}`)
	mdBold2Star  = regexp.MustCompile(`\*{2}(.+?)\*{2}`)
	mdItalicStar = regexp.MustCompile(`\*(.+?)\*`)
	mdBold3Under = regexp.MustCompile(`_{3}(.+?)_{3}`)
	mdBold2Under = regexp.MustCompile(`_{2}(.+?)_{2}`)
	mdItalicUndr = regexp.MustCompile(`_(.+?)_`)
	mdStrike     = regexp.MustCompile(`~~(.+?)~~`)
	mdCode       = regexp.MustCompile("```[^`]*```|`([^`]+)`")
	mdLink       = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	mdImage      = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	mdBlockquote = regexp.MustCompile(`(?m)^>\s?`)
	mdHR         = regexp.MustCompile(`(?m)^[-*_]{3,}\s*$`)
	mdListBullet = regexp.MustCompile(`(?m)^[ \t]*[*+-]\s+`)
	mdListNum    = regexp.MustCompile(`(?m)^[ \t]*\d+\.\s+`)
)

// StripMarkdown removes common Markdown formatting from s and returns
// plain text. It handles headings, bold, italic, strikethrough, inline
// code, code fences, links, images, blockquotes, horizontal rules, and
// list markers.
func StripMarkdown(s string) string {
	s = mdImage.ReplaceAllString(s, "$1")
	s = mdLink.ReplaceAllString(s, "$1")
	s = mdCode.ReplaceAllString(s, "$1")
	s = mdBold3Star.ReplaceAllString(s, "$1")
	s = mdBold2Star.ReplaceAllString(s, "$1")
	s = mdItalicStar.ReplaceAllString(s, "$1")
	s = mdBold3Under.ReplaceAllString(s, "$1")
	s = mdBold2Under.ReplaceAllString(s, "$1")
	s = mdItalicUndr.ReplaceAllString(s, "$1")
	s = mdStrike.ReplaceAllString(s, "$1")
	s = mdHeading.ReplaceAllString(s, "")
	s = mdBlockquote.ReplaceAllString(s, "")
	s = mdHR.ReplaceAllString(s, "")
	s = mdListBullet.ReplaceAllString(s, "")
	s = mdListNum.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// Redaction patterns for sensitive data.
var (
	redactSSN   = regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	redactCC    = regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	redactEmail = regexp.MustCompile(`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}\b`)
)

// Redact replaces common sensitive patterns (SSN, credit card numbers,
// email addresses) in s with "[REDACTED]".
func Redact(s string) string {
	s = redactSSN.ReplaceAllString(s, "[REDACTED]")
	s = redactCC.ReplaceAllString(s, "[REDACTED]")
	s = redactEmail.ReplaceAllString(s, "[REDACTED]")
	return s
}
