// Package cors provides configurable Cross-Origin Resource Sharing middleware
// for net/http servers.
//
// Call [Middleware] with a [Config] to obtain an http.Handler wrapper that sets
// the appropriate CORS response headers.  The middleware handles preflight
// OPTIONS requests automatically, returning 204 No Content with the required
// headers and stopping the handler chain.
//
// When [Config.AllowCredentials] is true and the origin list contains the
// wildcard "*", the middleware echoes back the request's Origin header instead
// of a literal "*" to comply with the CORS specification.
package cors
