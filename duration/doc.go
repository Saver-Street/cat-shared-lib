// Package duration provides human-readable formatting for time.Duration values.
//
// The package offers two formatting styles:
//   - Human: full format showing all significant units ("2h 30m 10s")
//   - Short: abbreviated format showing at most two units ("2h 30m")
//
// Convenience functions Since and Until format the duration relative to the
// current time, and Round / Truncate wrap the standard library helpers with
// a zero-precision guard.
package duration
