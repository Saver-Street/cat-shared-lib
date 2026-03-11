package batch

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Chunk splits items into sub-slices of at most size elements. The last chunk
// may contain fewer elements. Chunk panics if size < 1.
func Chunk[T any](items []T, size int) [][]T {
	if size < 1 {
		panic("batch.Chunk: size must be >= 1")
	}
	if len(items) == 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(items)+size-1)/size)
	for i := 0; i < len(items); i += size {
		end := i + size
		if end > len(items) {
			end = len(items)
		}
		chunks = append(chunks, items[i:end])
	}
	return chunks
}

// Process applies fn to each batch of items sequentially. It stops at the
// first error returned by fn. The context is passed to fn for cancellation.
func Process[T any](ctx context.Context, items []T, batchSize int, fn func(ctx context.Context, batch []T) error) error {
	for _, chunk := range Chunk(items, batchSize) {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("batch.Process: %w", err)
		}
		if err := fn(ctx, chunk); err != nil {
			return err
		}
	}
	return nil
}

// ProcessConcurrent applies fn to batches of items with at most maxWorkers
// goroutines running concurrently. It collects all errors and returns them
// as a joined error. If maxWorkers < 1 it defaults to 1.
func ProcessConcurrent[T any](ctx context.Context, items []T, batchSize, maxWorkers int, fn func(ctx context.Context, batch []T) error) error {
	if maxWorkers < 1 {
		maxWorkers = 1
	}

	chunks := Chunk(items, batchSize)
	if len(chunks) == 0 {
		return nil
	}

	sem := make(chan struct{}, maxWorkers)
	var (
		mu   sync.Mutex
		errs []error
	)

	var wg sync.WaitGroup
	for _, chunk := range chunks {
		if ctx.Err() != nil {
			break
		}

		wg.Add(1)
		sem <- struct{}{} // acquire slot
		go func(batch []T) {
			defer wg.Done()
			defer func() { <-sem }() // release slot

			if err := fn(ctx, batch); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(chunk)
	}

	wg.Wait()
	return errors.Join(errs...)
}

// Map applies fn to each item in items, collecting results in batches. It
// processes items sequentially in batches of batchSize. If fn returns an
// error, processing stops and the partial results are discarded.
func Map[T, U any](ctx context.Context, items []T, batchSize int, fn func(ctx context.Context, item T) (U, error)) ([]U, error) {
	results := make([]U, 0, len(items))
	for _, chunk := range Chunk(items, batchSize) {
		if err := ctx.Err(); err != nil {
			return nil, fmt.Errorf("batch.Map: %w", err)
		}
		for _, item := range chunk {
			result, err := fn(ctx, item)
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
	}
	return results, nil
}
