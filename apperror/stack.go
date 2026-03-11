package apperror

import (
	"fmt"
	"runtime"
	"strings"
)

// Frame represents a single stack frame.
type Frame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// String returns a human-readable representation of the frame.
func (f Frame) String() string {
	return fmt.Sprintf("%s\n\t%s:%d", f.Function, f.File, f.Line)
}

// StackTrace is an ordered list of stack frames from innermost to outermost.
type StackTrace []Frame

// String returns a multi-line formatted stack trace.
func (st StackTrace) String() string {
	if len(st) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, f := range st {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(f.String())
	}
	return sb.String()
}

// captureStack captures the current call stack, skipping the given number of
// frames (in addition to captureStack itself). maxDepth limits the total
// frames captured.
func captureStack(skip, maxDepth int) StackTrace {
	pcs := make([]uintptr, maxDepth)
	// +2: skip captureStack itself + runtime.Callers
	n := runtime.Callers(skip+2, pcs)
	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])
	var st StackTrace
	for {
		frame, more := frames.Next()
		st = append(st, Frame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}
	return st
}

// WithStack returns a copy of err with a captured stack trace.
// If err is nil, returns nil. If err is an *Error, the stack is attached
// to the copy. If err is a different type, it is wrapped in an Internal error.
func WithStack(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if e, ok := err.(*Error); ok {
		// Copy to avoid mutating the original
		cp := *e
		cp.Stack = captureStack(1, 32)
		return &cp
	}
	appErr = InternalWrap("unexpected error", err)
	appErr.Stack = captureStack(1, 32)
	return appErr
}

// HasStack reports whether the error has a captured stack trace.
func HasStack(err error) bool {
	if e, ok := err.(*Error); ok {
		return len(e.Stack) > 0
	}
	return false
}
