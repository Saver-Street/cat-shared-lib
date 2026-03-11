// Package request provides HTTP request parsing helpers for extracting and
// validating query parameters, URL path parameters, pagination, and sorting.
//
// [ParsePagination] extracts page/limit values from query strings and returns
// a [Pagination] struct with computed offset.  [RequireURLParam] and
// [RequireQueryParam] (plus their Int variants) return an error when a
// required parameter is missing.  [OptionalQueryParam] and [OptionalQueryInt]
// fall back to caller-supplied defaults.
//
// [ParseBoolParam] interprets common boolean representations.
// [ParseCommaSeparated] and [ParseCommaSeparatedInts] split multi-value query
// parameters.  [ParseSortOrder] validates sort field and direction against an
// allow-list.
package request
