// Package encoding provides base62 and hex encoding helpers for URL-safe,
// compact representations of binary data.
package encoding

import (
	"errors"
	"math/big"
	"strings"
)

const base62Alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

var (
	// ErrInvalidBase62 is returned when a string contains invalid base62 characters.
	ErrInvalidBase62 = errors.New("encoding: invalid base62 character")

	base62Big = big.NewInt(62)
)

// Base62Encode encodes raw bytes into a base62 string. The encoding is
// compact and URL-safe (no special characters).
func Base62Encode(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	n := new(big.Int).SetBytes(data)
	if n.Sign() == 0 {
		// Preserve leading zero bytes
		return strings.Repeat(string(base62Alphabet[0]), len(data))
	}

	var b strings.Builder
	b.Grow(len(data) * 2)
	mod := new(big.Int)
	for n.Sign() > 0 {
		n.DivMod(n, base62Big, mod)
		b.WriteByte(base62Alphabet[mod.Int64()])
	}

	// Reverse the string
	result := []byte(b.String())
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}

// Base62Decode decodes a base62 string back to raw bytes.
func Base62Decode(s string) ([]byte, error) {
	if s == "" {
		return nil, nil
	}
	n := new(big.Int)
	for _, c := range s {
		idx := strings.IndexRune(base62Alphabet, c)
		if idx < 0 {
			return nil, ErrInvalidBase62
		}
		n.Mul(n, base62Big)
		n.Add(n, big.NewInt(int64(idx)))
	}
	return n.Bytes(), nil
}

// Base62EncodeUint64 encodes a uint64 as a compact base62 string.
func Base62EncodeUint64(v uint64) string {
	if v == 0 {
		return string(base62Alphabet[0])
	}
	var buf [11]byte // max 11 chars for uint64
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = base62Alphabet[v%62]
		v /= 62
	}
	return string(buf[i:])
}

// Base62DecodeUint64 decodes a base62 string to a uint64.
func Base62DecodeUint64(s string) (uint64, error) {
	if s == "" {
		return 0, ErrInvalidBase62
	}
	var result uint64
	for _, c := range s {
		idx := strings.IndexRune(base62Alphabet, c)
		if idx < 0 {
			return 0, ErrInvalidBase62
		}
		prev := result
		result = result*62 + uint64(idx)
		if result < prev {
			return 0, errors.New("encoding: base62 overflow")
		}
	}
	return result, nil
}
