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
