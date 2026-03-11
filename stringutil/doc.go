// Package stringutil provides common string manipulation utilities
// that complement the standard library and the sanitize package.
//
// All functions are Unicode-aware, operating on runes rather than bytes.
// The package includes case-conversion helpers (ToKebabCase, ToPascalCase),
// padding (PadLeft, PadRight), content inspection (IsBlank, CountWords),
// rune-aware reversal (Reverse), and word-boundary line wrapping (WordWrap).
package stringutil
