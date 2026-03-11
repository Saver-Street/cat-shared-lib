// Package validation provides input validators for common formats and string
// constraints, returning user-facing [ValidationError] values.
//
// Format validators [Email], [UUID], [Phone], and [URL] check that a field
// value matches the expected pattern.  [Required] rejects empty strings, while
// [MinLength] and [MaxLength] enforce rune-count bounds.  [OneOf] ensures a
// value is among a set of allowed choices.
//
// Date validators [Date], [DateBefore], [DateAfter], [DateRange], [FutureDate],
// and [PastDate] validate time strings against Go time layouts and enforce
// temporal bounds.
//
// [Collect] gathers multiple validation results into a single error slice,
// filtering out nil entries, which simplifies validating an entire struct in
// one pass.
package validation
