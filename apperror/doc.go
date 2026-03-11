// Package apperror provides standardized application errors with HTTP status
// codes and machine-readable error codes for consistent error handling across
// microservices.
//
// Each error carries an HTTP status, a machine-readable [Code] such as
// [CodeNotFound] or [CodeValidation], and a human-readable message.
// Convenience constructors like [NotFound], [BadRequest], and [Internal] cover
// the most common cases, while [Wrap] and [InternalWrap] attach an underlying
// cause for use with [errors.Is] and [errors.As].
//
// Use [HTTPStatus] to extract the status from any error (defaults to 500) and
// [IsCode] to match on a specific error code without type-asserting.
package apperror
