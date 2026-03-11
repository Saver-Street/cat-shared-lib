// Package errutil provides generic error handling utilities that complement
// the apperror package. While apperror focuses on HTTP-specific structured
// errors, errutil offers general-purpose helpers for combining, wrapping,
// recovering from, and inspecting errors.
//
// Key functions:
//
//   - [Combine]: merge multiple errors into one (nil-safe)
//   - [Must], [MustOK]: panic-on-error wrappers for init and tests
//   - [Ignore]: discard error, return value only
//   - [Is], [As]: generic error matching (wraps [errors.Is] / [errors.As])
//   - [Wrap], [Wrapf]: add contextual messages to errors (nil-safe)
//   - [New]: shorthand for [fmt.Errorf]
//   - [Recover], [RecoverFunc]: convert panics to errors
//   - [Messages]: extract error strings from joined errors
package errutil
