// Package config provides helpers for loading typed configuration values from
// environment variables with sensible defaults and validation.
//
// Functions such as [String], [Int], [Bool], [Duration], and [StringSlice]
// return the environment variable's value converted to the requested type, or
// the supplied default when the variable is unset.  [MustString] and [MustInt]
// panic if the variable is missing, which is useful for values that must be
// present at startup.
//
// Call [Validate] with a list of required variable names to check them all at
// once; it returns a single error listing every missing variable.
//
// [Lookup] returns a value and boolean like map access.  [ValidateAll] ensures
// all keys in a group are set, [ValidateAny] requires at least one, and
// [FeatureEnabled] checks for truthy values (1, true, yes, on).
package config
