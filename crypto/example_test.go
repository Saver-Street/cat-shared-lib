package crypto_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/crypto"
)

func ExampleGenerateToken() {
	token, err := crypto.GenerateToken(32)
	fmt.Println(err)
	fmt.Println(len(token) > 0)
	// Output:
	// <nil>
	// true
}

func ExampleHMACSHA256() {
	key := []byte("secret-key")
	message := []byte("hello world")

	sig := crypto.HMACSHA256(key, message)
	fmt.Println(crypto.VerifyHMACSHA256(key, message, sig))
	fmt.Println(crypto.VerifyHMACSHA256(key, []byte("tampered"), sig))
	// Output:
	// true
	// false
}

func ExampleEqual() {
	fmt.Println(crypto.Equal("token-abc", "token-abc"))
	fmt.Println(crypto.Equal("token-abc", "token-xyz"))
	// Output:
	// true
	// false
}

func ExampleHashPassword() {
	hash, err := crypto.HashPassword("my-secure-password")
	fmt.Println(err)
	fmt.Println(len(hash) > 0)
	// Output:
	// <nil>
	// true
}

func ExampleCheckPassword() {
	hash, _ := crypto.HashPassword("correct-password")
	fmt.Println(crypto.CheckPassword("correct-password", hash))
	fmt.Println(crypto.CheckPassword("wrong-password", hash) != nil)
	// Output:
	// <nil>
	// true
}

func ExampleGenerateHexToken() {
	token, err := crypto.GenerateHexToken(16)
	fmt.Println(err)
	fmt.Println(len(token))
	// Output:
	// <nil>
	// 32
}

func ExampleVerifyHMACSHA256() {
	key := []byte("secret")
	msg := []byte("data")
	sig := crypto.HMACSHA256(key, msg)

	fmt.Println(crypto.VerifyHMACSHA256(key, msg, sig))
	fmt.Println(crypto.VerifyHMACSHA256(key, msg, "badsig"))
	// Output:
	// true
	// false
}

func ExampleNeedsRehash() {
	hash, _ := crypto.HashPasswordWithCost("pw", 10)
	fmt.Println(crypto.NeedsRehash(hash, 10))
	fmt.Println(crypto.NeedsRehash(hash, 12))
	// Output:
	// false
	// true
}

func ExampleEncrypt() {
	key := make([]byte, 32) // use a real key in production
	plaintext := []byte("sensitive data")

	ciphertext, err := crypto.Encrypt(key, plaintext)
	fmt.Println(err)
	fmt.Println(len(ciphertext) > len(plaintext))

	decrypted, err := crypto.Decrypt(key, ciphertext)
	fmt.Println(err)
	fmt.Println(string(decrypted))
	// Output:
	// <nil>
	// true
	// <nil>
	// sensitive data
}

func ExampleEncryptString() {
	key := make([]byte, 32) // use a real key in production

	encoded, err := crypto.EncryptString(key, "hello, world!")
	fmt.Println(err)
	fmt.Println(len(encoded) > 0)

	decoded, err := crypto.DecryptString(key, encoded)
	fmt.Println(err)
	fmt.Println(decoded)
	// Output:
	// <nil>
	// true
	// <nil>
	// hello, world!
}

func ExampleEqualBytes() {
	a := []byte("token-abc")
	b := []byte("token-abc")
	c := []byte("token-xyz")
	fmt.Println(crypto.EqualBytes(a, b))
	fmt.Println(crypto.EqualBytes(a, c))
	// Output:
	// true
	// false
}
