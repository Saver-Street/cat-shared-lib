package randutil_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/randutil"
)

func ExamplePick() {
	colors := []string{"red", "green", "blue", "yellow"}
	c := randutil.Pick(colors)
	fmt.Println(len(c) > 0)
	// Output: true
}

func ExampleShuffle() {
	nums := []int{1, 2, 3, 4, 5}
	shuffled := randutil.Shuffle(nums)
	fmt.Println(len(shuffled))
	// Output: 5
}

func ExampleAlphaNum() {
	s := randutil.AlphaNum(12)
	fmt.Println(len(s))
	// Output: 12
}
