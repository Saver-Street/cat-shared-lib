package types

import "fmt"

const bitsPerWord = 64

// BitSet is a compact set of non-negative integers backed by a slice of
// uint64 words. It supports standard set operations and is suitable for
// flags, bloom filters, and sparse integer sets.
type BitSet struct {
	words []uint64
	len   int
}

// NewBitSet creates a BitSet with capacity for at least size bits.
func NewBitSet(size int) *BitSet {
	n := (size + bitsPerWord - 1) / bitsPerWord
	if n == 0 {
		n = 1
	}
	return &BitSet{words: make([]uint64, n)}
}

// Set sets bit i to 1.
func (b *BitSet) Set(i int) {
	b.grow(i)
	if b.words[i/bitsPerWord]&(1<<(uint(i)%bitsPerWord)) == 0 {
		b.len++
	}
	b.words[i/bitsPerWord] |= 1 << (uint(i) % bitsPerWord)
}

// Clear sets bit i to 0.
func (b *BitSet) Clear(i int) {
	if i/bitsPerWord >= len(b.words) {
		return
	}
	if b.words[i/bitsPerWord]&(1<<(uint(i)%bitsPerWord)) != 0 {
		b.len--
	}
	b.words[i/bitsPerWord] &^= 1 << (uint(i) % bitsPerWord)
}

// Test reports whether bit i is set.
func (b *BitSet) Test(i int) bool {
	if i/bitsPerWord >= len(b.words) {
		return false
	}
	return b.words[i/bitsPerWord]&(1<<(uint(i)%bitsPerWord)) != 0
}

// Toggle flips bit i.
func (b *BitSet) Toggle(i int) {
	if b.Test(i) {
		b.Clear(i)
	} else {
		b.Set(i)
	}
}

// Len returns the number of set bits (population count).
func (b *BitSet) Len() int {
	return b.len
}

// Cap returns the current capacity in bits.
func (b *BitSet) Cap() int {
	return len(b.words) * bitsPerWord
}

// String returns a human-readable representation like "{0, 3, 7}".
func (b *BitSet) String() string {
	result := "{"
	first := true
	for i := range len(b.words) * bitsPerWord {
		if b.Test(i) {
			if !first {
				result += ", "
			}
			result += fmt.Sprintf("%d", i)
			first = false
		}
	}
	return result + "}"
}

// Union returns a new BitSet containing bits set in either b or other.
func (b *BitSet) Union(other *BitSet) *BitSet {
	maxLen := len(b.words)
	if len(other.words) > maxLen {
		maxLen = len(other.words)
	}
	result := &BitSet{words: make([]uint64, maxLen)}
	for i := range maxLen {
		var w1, w2 uint64
		if i < len(b.words) {
			w1 = b.words[i]
		}
		if i < len(other.words) {
			w2 = other.words[i]
		}
		result.words[i] = w1 | w2
	}
	result.recount()
	return result
}

// Intersect returns a new BitSet containing bits set in both b and other.
func (b *BitSet) Intersect(other *BitSet) *BitSet {
	minLen := len(b.words)
	if len(other.words) < minLen {
		minLen = len(other.words)
	}
	result := &BitSet{words: make([]uint64, minLen)}
	for i := range minLen {
		result.words[i] = b.words[i] & other.words[i]
	}
	result.recount()
	return result
}

// Diff returns a new BitSet containing bits set in b but not in other.
func (b *BitSet) Diff(other *BitSet) *BitSet {
	result := &BitSet{words: make([]uint64, len(b.words))}
	for i := range len(b.words) {
		var w2 uint64
		if i < len(other.words) {
			w2 = other.words[i]
		}
		result.words[i] = b.words[i] &^ w2
	}
	result.recount()
	return result
}

func (b *BitSet) grow(i int) {
	needed := i/bitsPerWord + 1
	if needed <= len(b.words) {
		return
	}
	newWords := make([]uint64, needed)
	copy(newWords, b.words)
	b.words = newWords
}

func (b *BitSet) recount() {
	b.len = 0
	for _, w := range b.words {
		b.len += popcount(w)
	}
}

func popcount(x uint64) int {
	// Kernighan's bit counting
	count := 0
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}
