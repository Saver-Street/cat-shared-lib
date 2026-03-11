// Package cache provides a thread-safe, generic in-memory LRU cache with
// per-entry TTL support and automatic background cleanup of expired entries.
//
// Create a cache with [New], supplying a [Config] that controls the maximum
// number of entries, the default TTL, and how often the background sweeper
// runs.  Use [Cache.Set] and [Cache.Get] for basic operations, or
// [Cache.SetWithTTL] to override the default TTL for individual entries.
// When the entry limit is reached the least-recently-used entry is evicted.
//
// Call [Cache.Stop] when the cache is no longer needed to release the
// background cleanup goroutine.
package cache
