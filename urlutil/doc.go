// Package urlutil provides convenience functions for URL manipulation.
//
// All functions accept raw URL strings and are safe for malformed input,
// returning the original string (or an empty string) when parsing fails.
//
// Functions include joining URL paths, setting/adding/removing query
// parameters, stripping query strings, extracting domains, and
// classifying URLs by scheme.
package urlutil
