package crypto

import "testing"

func FuzzHMACSHA256(f *testing.F) {
	f.Add([]byte("secret"), []byte("message"))
	f.Add([]byte(""), []byte(""))
	f.Add([]byte("key"), []byte(""))
	f.Add([]byte(""), []byte("data"))
	f.Fuzz(func(t *testing.T, key, message []byte) {
		sig := HMACSHA256(key, message)
		if sig == "" {
			t.Error("HMACSHA256 returned empty string")
		}
		if !VerifyHMACSHA256(key, message, sig) {
			t.Error("VerifyHMACSHA256 failed for valid signature")
		}
	})
}

func FuzzEqual(f *testing.F) {
	f.Add("abc", "abc")
	f.Add("abc", "xyz")
	f.Add("", "")
	f.Add("short", "longer-string")
	f.Fuzz(func(t *testing.T, a, b string) {
		result := Equal(a, b)
		expected := a == b
		if result != expected {
			t.Errorf("Equal(%q, %q) = %v, want %v", a, b, result, expected)
		}
	})
}

func FuzzEncryptDecrypt(f *testing.F) {
	f.Add([]byte("hello world"))
	f.Add([]byte(""))
	f.Add([]byte("short"))
	f.Add([]byte{0, 1, 2, 255, 254, 253})
	f.Fuzz(func(t *testing.T, plaintext []byte) {
		key := make([]byte, 32)
		for i := range key {
			key[i] = byte(i)
		}
		ct, err := Encrypt(key, plaintext)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		got, err := Decrypt(key, ct)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if len(plaintext) == 0 && len(got) == 0 {
			return
		}
		if string(got) != string(plaintext) {
			t.Errorf("roundtrip mismatch: got %q, want %q", got, plaintext)
		}
	})
}

func FuzzEqualBytes(f *testing.F) {
	f.Add([]byte("abc"), []byte("abc"))
	f.Add([]byte("abc"), []byte("xyz"))
	f.Add([]byte{}, []byte{})
	f.Add([]byte("short"), []byte("longer"))
	f.Fuzz(func(t *testing.T, a, b []byte) {
		result := EqualBytes(a, b)
		expected := len(a) == len(b)
		if expected {
			for i := range a {
				if a[i] != b[i] {
					expected = false
					break
				}
			}
		}
		if result != expected {
			t.Errorf("EqualBytes(%v, %v) = %v, want %v", a, b, result, expected)
		}
	})
}
