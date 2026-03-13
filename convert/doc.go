// Package convert provides safe type conversion functions with fallback defaults.
//
// Every conversion function accepts a fallback value that is returned when
// the input cannot be parsed. This eliminates the need for error handling
// at the call site for simple configuration or request-parameter parsing.
//
// The package also includes Must variants that panic on failure (useful in
// init functions or tests), and generic pointer helpers (Ptr, Deref, DerefOr).
//
// Example:
//
// port := convert.ToInt(os.Getenv("PORT"), 8080)
// debug := convert.ToBool(os.Getenv("DEBUG"), false)
// timeout := convert.ToDuration(os.Getenv("TIMEOUT"), 30*time.Second)
package convert
