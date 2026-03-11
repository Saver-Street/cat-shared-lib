// Package discovery provides a thread-safe, in-memory service registry with
// round-robin load balancing and health-status tracking.
//
// Create a registry with [NewRegistry] and register service instances via
// [Registry.Register].  Call [Registry.Resolve] to obtain a healthy instance
// using round-robin selection, or [Registry.ResolveHealthy] to retrieve all
// healthy instances for a given service name.
//
// Instance health is managed with [Registry.SetStatus] and the [Status]
// constants [StatusHealthy], [StatusUnhealthy], and [StatusDraining].
// [Registry.Heartbeat] refreshes an instance's last-seen timestamp, and
// [Registry.MarkStale] marks instances that have not reported within a TTL as
// unhealthy.  Use [WithOnInstanceStateChange] to receive callbacks when an
// instance's status changes.
package discovery
