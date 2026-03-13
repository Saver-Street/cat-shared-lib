// Package timeout provides utilities for running functions with time limits.
//
// The package offers four main functions:
//
//   - Do[T]: runs a function with a timeout and returns its result or a timeout error.
//   - DoSimple: a simplified variant of Do for functions that return only an error.
//   - After[T]: runs a function asynchronously and delivers the result on a channel.
//   - Race[T]: runs multiple functions concurrently and returns the first result.
//
// All functions are context-aware: child contexts are created where appropriate
// and cancelled on completion or timeout to prevent goroutine leaks.
package timeout
