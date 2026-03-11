// Package worker provides a bounded, context-aware worker pool for
// processing jobs concurrently.
//
// Create a pool with [New] specifying the maximum number of concurrent
// workers. Submit jobs with [Pool.Submit]. Each job receives a context
// that is cancelled when the pool is shut down.
//
//	pool := worker.New(4) // 4 concurrent workers
//	pool.Submit(func(ctx context.Context) error {
//	    return processItem(ctx, item)
//	})
//	errs := pool.Shutdown(ctx) // wait for all jobs, collect errors
//
// The pool collects errors returned by jobs. Call [Pool.Shutdown] to wait
// for all submitted jobs to complete and retrieve any errors.
package worker
