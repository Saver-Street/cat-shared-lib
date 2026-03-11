// Package cond provides generic conditional and comparison utility functions.
//
// It includes helpers such as Ternary (inline if/else), Coalesce (first
// non-zero value), Clamp (restrict to range), Min/Max (variadic extremes),
// Zero/IsZero (zero-value helpers), and Switch (pattern matching with Case
// structs).
//
// All functions are generic and type-safe, leveraging Go 1.18+ type
// parameters.
package cond
