// Package logging provides structured logging built on [log/slog] with
// JSON and text output, context propagation, and common attribute helpers.
//
// Create a logger with [New] and optional configuration via [WithWriter],
// [WithLevel], [WithFormat] ("json" or "text"), [WithSource], and [WithAttrs].
//
// Store and retrieve loggers in [context.Context] via [WithContext] and
// [FromContext].  [With] adds attributes to the logger in the context.
//
// Attribute helpers [RequestID], [UserID], [Service], [Version], [Component],
// [TraceID], and [Err] produce typed [slog.Attr] values for consistent field
// naming across Catherine microservices.
package logging
