package crypto

import "testing"

func BenchmarkHashPassword(b *testing.B) {
	// Use minimum cost for benchmarking to avoid slow runs.
	for b.Loop() {
		HashPasswordWithCost("benchmarkPassword123!", 4)
	}
}

func BenchmarkCheckPassword(b *testing.B) {
	hash, _ := HashPasswordWithCost("benchmarkPassword123!", 4)
	for b.Loop() {
		CheckPassword("benchmarkPassword123!", hash)
	}
}

func BenchmarkHMACSHA256(b *testing.B) {
	key := []byte("secret-key-for-benchmarking")
	msg := []byte("message to sign for benchmark testing")
	for b.Loop() {
		HMACSHA256(key, msg)
	}
}

func BenchmarkVerifyHMACSHA256(b *testing.B) {
	key := []byte("secret-key-for-benchmarking")
	msg := []byte("message to sign for benchmark testing")
	sig := HMACSHA256(key, msg)
	for b.Loop() {
		VerifyHMACSHA256(key, msg, sig)
	}
}

func BenchmarkEqual(b *testing.B) {
	a := "some-token-value-that-is-reasonably-long"
	c := "some-token-value-that-is-reasonably-long"
	for b.Loop() {
		Equal(a, c)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for b.Loop() {
		GenerateToken(32)
	}
}

func BenchmarkEncrypt(b *testing.B) {
	key := make([]byte, 32)
	plaintext := []byte("sensitive data that needs encryption for storage")
	for b.Loop() {
		Encrypt(key, plaintext)
	}
}

func BenchmarkDecrypt(b *testing.B) {
	key := make([]byte, 32)
	ct, _ := Encrypt(key, []byte("sensitive data that needs encryption for storage"))
	for b.Loop() {
		Decrypt(key, ct)
	}
}

func BenchmarkEqualBytes(b *testing.B) {
	a := []byte("some-token-value-that-is-reasonably-long")
	c := []byte("some-token-value-that-is-reasonably-long")
	for b.Loop() {
		EqualBytes(a, c)
	}
}
