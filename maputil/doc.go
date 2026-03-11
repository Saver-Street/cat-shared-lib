// Package maputil provides generic utility functions for working with Go maps.
//
// All functions are safe for nil maps and return new maps rather than
// modifying the input.
//
//	m := map[string]int{"a": 1, "b": 2, "c": 3}
//	keys := maputil.Keys(m)           // [a b c] (unordered)
//	sub  := maputil.Pick(m, "a", "c") // {a:1, c:3}
//	big  := maputil.Filter(m, func(k string, v int) bool { return v > 1 })
package maputil
