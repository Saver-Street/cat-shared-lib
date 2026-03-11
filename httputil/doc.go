// Package httputil provides common HTTP utility functions for request
// inspection, content type detection, authentication header extraction,
// and response helpers.
//
// It complements the existing request, response, and middleware packages
// by offering lightweight, stateless helper functions that do not depend
// on any particular router or framework.
//
// # Content Type Detection
//
// IsJSON, IsForm, and IsMultipart inspect the Content-Type header to
// determine the encoding of the request body.
//
// # Authentication Helpers
//
// BearerToken extracts a bearer token from the Authorization header.
// BasicAuth wraps the standard library BasicAuth for convenience.
//
// # Response Helpers
//
// WriteJSON and WriteError provide shorthand methods for writing JSON
// responses with appropriate headers and status codes.
package httputil
