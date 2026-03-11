package maputil

// Keys returns a slice of all keys in m. The order is not guaranteed.
func Keys[K comparable, V any](m map[K]V) []K {
	out := make([]K, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// Values returns a slice of all values in m. The order is not guaranteed.
func Values[K comparable, V any](m map[K]V) []V {
	out := make([]V, 0, len(m))
	for _, v := range m {
		out = append(out, v)
	}
	return out
}

// Merge combines multiple maps into one. Later maps overwrite earlier ones
// for duplicate keys.
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	out := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}

// Pick returns a new map containing only the specified keys.
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	out := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			out[k] = v
		}
	}
	return out
}

// Omit returns a new map excluding the specified keys.
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	exclude := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		exclude[k] = struct{}{}
	}
	out := make(map[K]V, len(m))
	for k, v := range m {
		if _, skip := exclude[k]; !skip {
			out[k] = v
		}
	}
	return out
}

// Filter returns a new map containing only the entries for which keep
// returns true.
func Filter[K comparable, V any](m map[K]V, keep func(K, V) bool) map[K]V {
	out := make(map[K]V)
	for k, v := range m {
		if keep(k, v) {
			out[k] = v
		}
	}
	return out
}

// MapValues applies fn to each value in m and returns a new map with the
// same keys and transformed values.
func MapValues[K comparable, V, U any](m map[K]V, fn func(V) U) map[K]U {
	out := make(map[K]U, len(m))
	for k, v := range m {
		out[k] = fn(v)
	}
	return out
}

// Invert swaps keys and values. If multiple keys have the same value, one
// wins (non-deterministic).
func Invert[K, V comparable](m map[K]V) map[V]K {
	out := make(map[V]K, len(m))
	for k, v := range m {
		out[v] = k
	}
	return out
}

// Equal reports whether two maps have the same keys and values.
func Equal[K, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		if vb, ok := b[k]; !ok || va != vb {
			return false
		}
	}
	return true
}
