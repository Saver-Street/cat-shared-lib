package duration_test

import (
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/duration"
)

func ExampleHuman() {
	fmt.Println(duration.Human(2*time.Hour + 30*time.Minute + 10*time.Second))
	fmt.Println(duration.Human(500 * time.Millisecond))
	// Output:
	// 2h 30m 10s
	// < 1s
}

func ExampleShort() {
	fmt.Println(duration.Short(2*time.Hour + 30*time.Minute + 45*time.Second))
	fmt.Println(duration.Short(750 * time.Millisecond))
	// Output:
	// 2h 30m
	// 750ms
}

func ExampleRound() {
	d := 2*time.Hour + 30*time.Minute
	fmt.Println(duration.Round(d, time.Hour))
	// Output:
	// 3h0m0s
}
