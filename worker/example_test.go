package worker_test

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/Saver-Street/cat-shared-lib/worker"
)

func ExamplePool() {
	pool := worker.New(2)

	var sum atomic.Int64
	for _, n := range []int{1, 2, 3, 4, 5} {
		pool.Submit(func(_ context.Context) error {
			sum.Add(int64(n))
			return nil
		})
	}

	errs := pool.Shutdown(context.Background())
	fmt.Printf("sum=%d errors=%d\n", sum.Load(), len(errs))
	// Output:
	// sum=15 errors=0
}
