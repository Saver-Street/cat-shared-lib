package sliceutil_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/sliceutil"
)

func ExamplePartition() {
	even, odd := sliceutil.Partition([]int{1, 2, 3, 4, 5}, func(n int) bool {
		return n%2 == 0
	})
	fmt.Println("even:", even)
	fmt.Println("odd:", odd)
	// Output:
	// even: [2 4]
	// odd: [1 3 5]
}

func ExampleReduce() {
	sum := sliceutil.Reduce([]int{1, 2, 3, 4}, 0, func(acc, n int) int {
		return acc + n
	})
	fmt.Println(sum)
	// Output:
	// 10
}

func ExampleFind() {
	v, ok := sliceutil.Find([]string{"apple", "banana", "cherry"}, func(s string) bool {
		return s == "banana"
	})
	fmt.Println(v, ok)
	// Output:
	// banana true
}

func ExampleAssociate() {
	type user struct {
		ID   int
		Name string
	}
	users := []user{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	m := sliceutil.Associate(users, func(u user) int { return u.ID })
	fmt.Println(m[1].Name)
	fmt.Println(m[2].Name)
	// Output:
	// Alice
	// Bob
}
