// Package response provides JSON response helpers for HTTP APIs, covering
// standard status codes, paginated results, request body decoding, and error
// envelopes.
//
// Convenience functions [OK], [Created], [Accepted], [NoContent], [BadRequest],
// [Unauthorized], [Forbidden], [NotFound], [Conflict], [UnprocessableEntity],
// [TooManyRequests], [ServiceUnavailable], [MethodNotAllowed], [Gone], and
// [GatewayTimeout] write a JSON response with the appropriate HTTP status.
// [InternalError] logs the underlying error before responding.
//
// [Paginated] writes a [PagedResult] envelope with page metadata.
// [PaginatedWithHeaders] additionally sets X-Total-Count and Link headers via
// [SetPaginationHeaders].
//
// [DecodeJSON] parses a request body into a struct, and [DecodeOrFail] does
// the same while automatically writing a 400 response on failure.
package response
