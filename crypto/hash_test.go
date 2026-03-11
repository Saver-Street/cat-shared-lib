package crypto

import (
	"bytes"
	"io"
	"testing"
)

func TestHashSHA512(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data []byte
		want string
	}{
		{"empty", []byte{}, "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"},
		{"hello", []byte("hello"), "9b71d224bd62f3785d96d46ad3ea3d73319bfbc2890caadae2dff72519673ca72323c3d99ba5c11d7c7acc6e14b8c5da0c4663475c2e5c3adef46f73bcdec043"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := HashSHA512(tt.data)
			if got != tt.want {
				t.Errorf("HashSHA512() = %s; want %s", got, tt.want)
			}
		})
	}
}

func TestHashReader(t *testing.T) {
	t.Parallel()
	t.Run("valid", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader([]byte("hello"))
		got, err := HashReader(r)
		if err != nil {
			t.Fatalf("HashReader() error = %v", err)
		}
		want := HashSHA512([]byte("hello"))
		if got != want {
			t.Errorf("HashReader() = %s; want %s", got, want)
		}
	})
	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		r := bytes.NewReader(nil)
		got, err := HashReader(r)
		if err != nil {
			t.Fatalf("HashReader() error = %v", err)
		}
		want := HashSHA512(nil)
		if got != want {
			t.Errorf("HashReader() = %s; want %s", got, want)
		}
	})
	t.Run("error", func(t *testing.T) {
		t.Parallel()
		r := &errReader{}
		_, err := HashReader(r)
		if err == nil {
			t.Error("HashReader() error = nil; want error")
		}
	})
}

type errReader struct{}

func (e *errReader) Read([]byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestHMACSHA512(t *testing.T) {
	t.Parallel()
	key := []byte("secret")
	msg := []byte("hello")
	sig := HMACSHA512(key, msg)
	if len(sig) != 128 {
		t.Errorf("HMACSHA512() length = %d; want 128", len(sig))
	}
	if sig2 := HMACSHA512(key, msg); sig != sig2 {
		t.Error("HMACSHA512() not deterministic")
	}
}

func TestVerifyHMACSHA512(t *testing.T) {
	t.Parallel()
	key := []byte("secret")
	msg := []byte("hello")
	sig := HMACSHA512(key, msg)

	if !VerifyHMACSHA512(key, msg, sig) {
		t.Error("VerifyHMACSHA512() = false; want true")
	}
	if VerifyHMACSHA512(key, msg, "invalid") {
		t.Error("VerifyHMACSHA512(invalid) = true; want false")
	}
	if VerifyHMACSHA512([]byte("wrong"), msg, sig) {
		t.Error("VerifyHMACSHA512(wrong key) = true; want false")
	}
}

func BenchmarkHashSHA512(b *testing.B) {
	data := []byte("benchmark data for hashing performance test")
	for range b.N {
		HashSHA512(data)
	}
}

func BenchmarkHMACSHA512(b *testing.B) {
	key := []byte("secret-key")
	msg := []byte("benchmark message data")
	for range b.N {
		HMACSHA512(key, msg)
	}
}

func FuzzHashSHA512(f *testing.F) {
	f.Add([]byte("hello"))
	f.Add([]byte{})
	f.Add([]byte("test data"))
	f.Fuzz(func(t *testing.T, data []byte) {
		result := HashSHA512(data)
		if len(result) != 128 {
			t.Errorf("HashSHA512() length = %d; want 128", len(result))
		}
	})
}
