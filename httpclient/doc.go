// Package httpclient provides a resilient HTTP client for service-to-service
// communication with automatic retries, circuit breaking, and JSON helpers.
//
// Create a client with [New] and functional [Option] values such as
// [WithTimeout], [WithRetries], [WithCircuitBreaker], [WithRequestHook], and
// [WithResponseHook].  The client retries failed requests using exponential
// backoff with jitter, and an optional circuit breaker (from the
// circuitbreaker package) can short-circuit calls to unhealthy upstreams.
//
// Convenience methods [Client.GetJSON], [Client.PostJSON], [Client.PutJSON],
// and [Client.DeleteJSON] marshal request bodies and unmarshal responses
// automatically.  Lower-level [Client.Get], [Client.Post], [Client.Put],
// [Client.Delete], and [Client.Do] return a [Response] for manual handling.
package httpclient
