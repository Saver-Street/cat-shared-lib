package errutil

import (
	"errors"
	"fmt"
	"strings"
)

// Combine merges multiple errors into one. Nil errors are ignored.
// Returns nil if all errors are nil.
func Combine(errs ...error) error {
	var nonNil []error
	for _, err := range errs {
		if err != nil {
			nonNil = append(nonNil, err)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	default:
		return errors.Join(nonNil...)
	}
}

// Must panics if err is non-nil, otherwise returns v.
// Useful for wrapping functions that return (T, error) where error is
// unexpected (e.g., in init or tests).
func Must[T any](v T, err error) T {
	if err != nil {
		panic(fmt.Sprintf("errutil.Must: %v", err))
	}
	return v
}

// MustOK panics if err is non-nil. For functions returning only error.
func MustOK(err error) {
	if err != nil {
		panic(fmt.Sprintf("errutil.MustOK: %v", err))
	}
}

// Ignore discards the error and returns only the value.
// Use sparingly; prefer explicit error handling.
func Ignore[T any](v T, _ error) T {
	return v
}

// Is reports whether any error in the chain matches target.
// Shorthand for errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in the chain that matches target.
// Shorthand for errors.As.
func As[T error](err error) (T, bool) {
	var target T
	ok := errors.As(err, &target)
	return target, ok
}

// Wrap wraps err with a message prefix. Returns nil if err is nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf wraps err with a formatted message prefix. Returns nil if err is nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// New creates a new error with a formatted message.
// Shorthand for fmt.Errorf.
func New(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// Recover calls fn and recovers from any panic, returning it as an error.
func Recover(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered panic: %v", r)
			}
		}
	}()
	fn()
	return nil
}

// RecoverFunc calls fn and recovers from any panic, returning it as an error.
// For functions that return a value.
func RecoverFunc[T any](fn func() T) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("recovered panic: %v", r)
			}
		}
	}()
	return fn(), nil
}

// Messages extracts error messages from a joined error into a slice.
// For a single error, returns a slice with one message.
func Messages(err error) []string {
	if err == nil {
		return nil
	}
	parts := strings.Split(err.Error(), "\n")
	var msgs []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			msgs = append(msgs, p)
		}
	}
	return msgs
}
