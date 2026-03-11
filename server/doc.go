// Package server provides a standard HTTP server with graceful shutdown for
// Catherine microservices.
//
// Call [ListenAndServe] with a [Config] specifying the address, handler, and
// timeouts for reads, writes, headers, and idle connections.  The server
// listens for SIGINT and SIGTERM, then drains in-flight requests within a
// configurable shutdown timeout before exiting.  Optional cleanup functions
// passed to [ListenAndServe] are executed during the shutdown sequence.
package server
