package ratelimit_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/ratelimit"
)

func ExampleLimiter_Allow() {
	l := ratelimit.New(ratelimit.Config{Rate: 5, Burst: 5})
	defer l.Stop()

	// First request within burst is allowed.
	fmt.Println(l.Allow("user-1"))
	// Output:
	// true
}

func ExampleLimiter_AllowN() {
	l := ratelimit.New(ratelimit.Config{Rate: 10, Burst: 10})
	defer l.Stop()

	fmt.Println(l.AllowN("user-1", 5)) // consumes 5 of 10
	fmt.Println(l.AllowN("user-1", 5)) // consumes remaining 5
	fmt.Println(l.AllowN("user-1", 1)) // over limit
	// Output:
	// true
	// true
	// false
}
