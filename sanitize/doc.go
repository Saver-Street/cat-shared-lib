// Package sanitize provides input sanitization and normalization utilities for
// filenames, text, emails, HTML, null-pointer handling, and PostgreSQL error
// detection.
//
// [DocFilename] and [TruncateFilename] clean and shorten file names for safe
// storage.  [MaxLength] truncates a string to a maximum rune count.
// [SanitizeEmail] lowercases and trims email addresses.  [SanitizePhone] strips
// non-digit characters from phone numbers.  [SanitizeURL] normalizes scheme and
// host.  [SanitizeName] title-cases and collapses whitespace in human names.  [StripHTML] removes
// HTML tags from a string.
//
// [NilIfEmpty] and [TrimAndNilIfEmpty] convert empty strings to nil pointers,
// useful for nullable database columns.  The generic [Deref] safely
// dereferences a pointer with a fallback default.
//
// [IsDuplicateKey] and [IsDatabaseError] inspect PostgreSQL errors by their
// SQLSTATE code, simplifying constraint-violation handling.
package sanitize
