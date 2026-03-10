package circuitbreaker_test

import (
	"errors"
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/circuitbreaker"
)

func ExampleNew() {
	cb := circuitbreaker.New("payment-api",
		circuitbreaker.WithFailureThreshold(3),
		circuitbreaker.WithResetTimeout(30*time.Second),
	)
	fmt.Println(cb.Name())
	fmt.Println(cb.State())
	// Output:
	// payment-api
	// closed
}

func ExampleBreaker_Execute() {
	cb := circuitbreaker.New("example",
		circuitbreaker.WithFailureThreshold(2),
	)

	// Successful call
	err := cb.Execute(func() error {
		return nil
	})
	fmt.Println("success:", err)

	// Failures that trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return errors.New("fail") })

	// Now the circuit is open
	err = cb.Execute(func() error { return nil })
	fmt.Println("open:", err)

	// Output:
	// success: <nil>
	// open: circuitbreaker: circuit is open
}

func ExampleBreaker_Reset() {
	cb := circuitbreaker.New("example", circuitbreaker.WithFailureThreshold(1))
	_ = cb.Execute(func() error { return errors.New("fail") })
	fmt.Println("before reset:", cb.State())

	cb.Reset()
	fmt.Println("after reset:", cb.State())
	// Output:
	// before reset: open
	// after reset: closed
}
