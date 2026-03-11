// Package health provides standardized health check handlers for Catherine
// microservices, supporting concurrent checks with a configurable timeout.
//
// Use [NewChecker] to create a [Checker] and register named check functions.
// Call [Handler] to obtain an http.HandlerFunc that runs all checks
// concurrently (with a 5-second timeout) and returns a JSON [Status] with
// state "ok" or "degraded".
//
// Built-in check constructors cover common dependencies:
//   - [DBChecker] verifies database connectivity via a [Pinger] interface.
//   - [HTTPChecker] probes an upstream HTTP endpoint.
//   - [ServiceDiscoveryChecker] validates that a service has healthy instances
//     in a discovery registry.
//   - [AggregateChecker] combines multiple checks into one.
//
// [Status.IsHealthy] and [Status.HasErrors] provide convenient result
// inspection.
package health
