package batch_test

import (
	"context"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/batch"
)

func ExampleChunk() {
	chunks := batch.Chunk([]int{1, 2, 3, 4, 5}, 2)
	for _, ch := range chunks {
		fmt.Println(ch)
	}
	// Output:
	// [1 2]
	// [3 4]
	// [5]
}

func ExampleProcess() {
	items := []string{"a", "b", "c", "d", "e"}
	err := batch.Process(context.Background(), items, 2, func(_ context.Context, b []string) error {
		fmt.Println("batch:", b)
		return nil
	})
	if err != nil {
		fmt.Println("error:", err)
	}
	// Output:
	// batch: [a b]
	// batch: [c d]
	// batch: [e]
}

func ExampleMap() {
	items := []int{1, 2, 3, 4, 5}
	results, err := batch.Map(context.Background(), items, 3, func(_ context.Context, item int) (string, error) {
		return fmt.Sprintf("item-%d", item), nil
	})
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(results)
	// Output:
	// [item-1 item-2 item-3 item-4 item-5]
}
