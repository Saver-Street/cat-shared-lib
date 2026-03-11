package encoding

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestBase62EncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()
	inputs := [][]byte{
		{0x01},
		{0xFF},
		{0xDE, 0xAD, 0xBE, 0xEF},
		[]byte("Hello, World!"),
		{0x00, 0x01, 0x02},
	}
	for _, input := range inputs {
		encoded := Base62Encode(input)
		if encoded == "" {
			t.Errorf("Base62Encode(%x) returned empty", input)
			continue
		}
		for _, c := range encoded {
			if !strings.ContainsRune(base62Alphabet, c) {
				t.Errorf("Base62Encode(%x) contains invalid char %c", input, c)
			}
		}
		decoded, err := Base62Decode(encoded)
		if err != nil {
			t.Errorf("Base62Decode(%q): %v", encoded, err)
			continue
		}
		// big.Int drops leading zeros, so compare trimmed
		inputTrimmed := input
		for len(inputTrimmed) > 0 && inputTrimmed[0] == 0 {
			inputTrimmed = inputTrimmed[1:]
		}
		if !bytes.Equal(decoded, inputTrimmed) {
			t.Errorf("roundtrip %x → %q → %x", input, encoded, decoded)
		}
	}
}

func TestBase62EncodeEmpty(t *testing.T) {
	t.Parallel()
	if s := Base62Encode(nil); s != "" {
		t.Errorf("Base62Encode(nil) = %q; want empty", s)
	}
}

func TestBase62EncodeAllZeros(t *testing.T) {
	t.Parallel()
	s := Base62Encode([]byte{0, 0, 0})
	if s != "000" {
		t.Errorf("Base62Encode([0,0,0]) = %q; want 000", s)
	}
}

func TestBase62DecodeEmpty(t *testing.T) {
	t.Parallel()
	b, err := Base62Decode("")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if b != nil {
		t.Errorf("Base62Decode('') = %v; want nil", b)
	}
}

func TestBase62DecodeInvalid(t *testing.T) {
	t.Parallel()
	_, err := Base62Decode("abc!def")
	if !errors.Is(err, ErrInvalidBase62) {
		t.Errorf("err = %v; want ErrInvalidBase62", err)
	}
}

func TestBase62EncodeUint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input uint64
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{61, "z"},
		{62, "10"},
		{1000, "G8"},
	}
	for _, tt := range tests {
		got := Base62EncodeUint64(tt.input)
		if got != tt.want {
			t.Errorf("Base62EncodeUint64(%d) = %q; want %q", tt.input, got, tt.want)
		}
	}
}

func TestBase62DecodeUint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  uint64
	}{
		{"0", 0},
		{"1", 1},
		{"z", 61},
		{"10", 62},
		{"G8", 1000},
	}
	for _, tt := range tests {
		got, err := Base62DecodeUint64(tt.input)
		if err != nil {
			t.Errorf("Base62DecodeUint64(%q): %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Base62DecodeUint64(%q) = %d; want %d", tt.input, got, tt.want)
		}
	}
}

func TestBase62DecodeUint64Empty(t *testing.T) {
	t.Parallel()
	_, err := Base62DecodeUint64("")
	if !errors.Is(err, ErrInvalidBase62) {
		t.Errorf("err = %v; want ErrInvalidBase62", err)
	}
}

func TestBase62DecodeUint64Invalid(t *testing.T) {
	t.Parallel()
	_, err := Base62DecodeUint64("abc!xyz")
	if !errors.Is(err, ErrInvalidBase62) {
		t.Errorf("err = %v; want ErrInvalidBase62", err)
	}
}

func TestBase62DecodeUint64Overflow(t *testing.T) {
	t.Parallel()
	// This string represents a value larger than uint64 max
	_, err := Base62DecodeUint64("zzzzzzzzzzzzz")
	if err == nil {
		t.Error("expected overflow error")
	}
}

func TestBase62Uint64RoundTrip(t *testing.T) {
	t.Parallel()
	values := []uint64{0, 1, 42, 1000, 1<<32 - 1, 1<<63 - 1}
	for _, v := range values {
		s := Base62EncodeUint64(v)
		got, err := Base62DecodeUint64(s)
		if err != nil {
			t.Errorf("roundtrip %d: decode error: %v", v, err)
			continue
		}
		if got != v {
			t.Errorf("roundtrip %d → %q → %d", v, s, got)
		}
	}
}

func BenchmarkBase62Encode(b *testing.B) {
	data := []byte("Hello, World!")
	for range b.N {
		Base62Encode(data)
	}
}

func BenchmarkBase62EncodeUint64(b *testing.B) {
	for range b.N {
		Base62EncodeUint64(123456789)
	}
}

func FuzzBase62RoundTrip(f *testing.F) {
	f.Add([]byte("test"))
	f.Add([]byte{0xFF, 0xFE})
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		encoded := Base62Encode(data)
		if len(data) == 0 && encoded != "" {
			t.Error("empty input should give empty output")
		}
		if encoded == "" {
			return
		}
		_, err := Base62Decode(encoded)
		if err != nil {
			t.Errorf("decode error on valid encode: %v", err)
		}
	})
}

func FuzzBase62Uint64(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(42))
	f.Add(uint64(1<<63 - 1))
	f.Fuzz(func(t *testing.T, v uint64) {
		s := Base62EncodeUint64(v)
		got, err := Base62DecodeUint64(s)
		if err != nil {
			t.Fatalf("decode %q: %v", s, err)
		}
		if got != v {
			t.Errorf("roundtrip %d → %q → %d", v, s, got)
		}
	})
}
