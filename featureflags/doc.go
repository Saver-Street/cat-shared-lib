// Package featureflags provides an environment-variable-backed feature flag
// manager with type-safe accessors for boolean, integer, float, and list
// values.
//
// Create a [Manager] with [NewManager] and register flags via
// [Manager.Register], specifying a name, default value, and description.  Each
// flag resolves to the environment variable {Prefix}{NAME} (default prefix
// "FEATURE_").  Query flags with [Manager.Enabled] (truthy: "1", "true",
// "yes", "on"), [Manager.IntValue], [Manager.Float64Value], or
// [Manager.ListValue].
//
// Call [Manager.All] for a snapshot of all registered flags and their current
// values, or [Manager.AllFlags] for the full [Flag] definitions.
package featureflags
