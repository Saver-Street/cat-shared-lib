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
