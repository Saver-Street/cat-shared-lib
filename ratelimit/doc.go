// Package ratelimit provides a standalone, thread-safe, per-key token bucket
// rate limiter with automatic cleanup of idle entries.
//
// Create a [Limiter] with [New] and a [Config] specifying the sustained rate,
// burst size, cleanup interval, and maximum idle time.  Call [Limiter.Allow]
// to check whether a single event for a given key should be permitted, or
// [Limiter.AllowN] for batch checks.
//
// A background goroutine periodically removes entries that have been idle
// longer than [Config.MaxIdleTime].  Call [Limiter.Stop] when the limiter is
// no longer needed to release the cleanup goroutine.  [Limiter.Len] returns
// the current number of tracked keys.
package ratelimit
