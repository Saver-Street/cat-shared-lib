package randutil

import (
	"math/rand/v2"
)

// Int returns a random int in [min, max). Panics if min >= max.
func Int(min, max int) int {
	if min >= max {
		panic("randutil.Int: min must be less than max")
	}
	return min + rand.IntN(max-min)
}

// Float64 returns a random float64 in [min, max).
func Float64(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

// Bool returns a random boolean.
func Bool() bool {
	return rand.IntN(2) == 1
}

// Pick returns a random element from items. Panics if items is empty.
func Pick[T any](items []T) T {
	return items[rand.IntN(len(items))]
}

// Shuffle returns a new slice with elements in random order.
func Shuffle[T any](items []T) []T {
	out := make([]T, len(items))
	copy(out, items)
	rand.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

// Sample returns n random elements from items without replacement.
// Panics if n > len(items).
func Sample[T any](items []T, n int) []T {
	if n > len(items) {
		panic("randutil.Sample: n exceeds slice length")
	}
	shuffled := Shuffle(items)
	return shuffled[:n]
}

// String returns a random string of length n from the given alphabet.
func String(n int, alphabet string) string {
	runes := []rune(alphabet)
	out := make([]rune, n)
	for i := range out {
		out[i] = runes[rand.IntN(len(runes))]
	}
	return string(out)
}

// Hex returns a random hex string of length n (n hex characters).
func Hex(n int) string {
	const hex = "0123456789abcdef"
	return String(n, hex)
}

// Alpha returns a random alphabetic string of length n.
func Alpha(n int) string {
	const alpha = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return String(n, alpha)
}

// AlphaNum returns a random alphanumeric string of length n.
func AlphaNum(n int) string {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return String(n, alphanum)
}

// WeightedPick returns a random element chosen by weights.
// weights[i] is the relative probability of items[i].
// Panics if items and weights have different lengths or if all weights are 0.
func WeightedPick[T any](items []T, weights []float64) T {
	if len(items) != len(weights) {
		panic("randutil.WeightedPick: items and weights must have same length")
	}
	var total float64
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		panic("randutil.WeightedPick: total weight is zero")
	}
	r := rand.Float64() * total
	var cum float64
	for i, w := range weights {
		cum += w
		if r < cum {
			return items[i]
		}
	}
	return items[len(items)-1]
}
