package types

import (
	"hash/fnv"
	"math"
)

// BloomFilter is a space-efficient probabilistic set.
// It may return false positives but never false negatives.
type BloomFilter struct {
	bits []uint64
	size uint
	k    uint
}

// NewBloomFilter creates a BloomFilter optimised for the expected number
// of elements and desired false positive probability.
// Panics if n < 1 or fpRate is not in (0, 1).
func NewBloomFilter(n uint, fpRate float64) *BloomFilter {
	if n < 1 {
		panic("bloom: n must be >= 1")
	}
	if fpRate <= 0 || fpRate >= 1 {
		panic("bloom: fpRate must be in (0, 1)")
	}
	m := optimalM(n, fpRate)
	k := optimalK(m, n)
	return &BloomFilter{
		bits: make([]uint64, (m+63)/64),
		size: m,
		k:    k,
	}
}

// Add inserts data into the filter.
func (bf *BloomFilter) Add(data []byte) {
	h1, h2 := bf.hashes(data)
	for i := range bf.k {
		idx := (h1 + uint64(i)*h2) % uint64(bf.size)
		bf.bits[idx/64] |= 1 << (idx % 64)
	}
}

// Contains reports whether data might be in the filter.
// False positives are possible; false negatives are not.
func (bf *BloomFilter) Contains(data []byte) bool {
	h1, h2 := bf.hashes(data)
	for i := range bf.k {
		idx := (h1 + uint64(i)*h2) % uint64(bf.size)
		if bf.bits[idx/64]&(1<<(idx%64)) == 0 {
			return false
		}
	}
	return true
}

// AddString is a convenience wrapper around Add for string values.
func (bf *BloomFilter) AddString(s string) {
	bf.Add([]byte(s))
}

// ContainsString is a convenience wrapper around Contains for string values.
func (bf *BloomFilter) ContainsString(s string) bool {
	return bf.Contains([]byte(s))
}

// hashes returns two independent 64-bit hash values using FNV-1a.
func (bf *BloomFilter) hashes(data []byte) (uint64, uint64) {
	h1 := fnvHash64(data)
	// For h2, prepend a byte to get a different hash.
	data2 := make([]byte, len(data)+1)
	data2[0] = 0x9e // arbitrary seed byte
	copy(data2[1:], data)
	h2 := fnvHash64(data2)
	return h1, h2
}

func fnvHash64(data []byte) uint64 {
	h := fnv.New64a()
	h.Write(data)
	return h.Sum64()
}

func optimalM(n uint, fpRate float64) uint {
	m := -float64(n) * math.Log(fpRate) / (math.Log(2) * math.Log(2))
	return uint(math.Ceil(m))
}

func optimalK(m, n uint) uint {
	k := float64(m) / float64(n) * math.Log(2)
	result := uint(math.Round(k))
	if result < 1 {
		return 1
	}
	return result
}
