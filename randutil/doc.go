// Package randutil provides convenient random value generation utilities
// for non-cryptographic use cases such as testing, ID generation, and sampling.
//
// All functions use [math/rand/v2] which is safe for concurrent use and
// automatically seeded. For cryptographic randomness, use the crypto package
// or [crypto/rand] directly.
//
// Basic random values:
//
//	randutil.Int(1, 100)       // random int in [1, 100)
//	randutil.Float64(0, 1)     // random float64 in [0, 1)
//	randutil.Bool()            // random true or false
//
// Slice operations:
//
//	randutil.Pick(colors)          // random element
//	randutil.Shuffle(deck)         // new shuffled slice
//	randutil.Sample(names, 3)      // 3 random names without replacement
//	randutil.WeightedPick(xs, ws)  // weighted random selection
//
// String generators:
//
//	randutil.String(8, "abc")  // 8-char string from custom alphabet
//	randutil.Hex(16)           // 16-char hex string
//	randutil.Alpha(10)         // 10-char letter string
//	randutil.AlphaNum(12)      // 12-char alphanumeric string
package randutil
