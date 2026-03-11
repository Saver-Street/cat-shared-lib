// Package batch provides helpers for processing slices of items in fixed-size
// chunks, optionally with bounded concurrency.
//
// [Process] applies a function to every item in a slice, processing them in
// batches of a configurable size.  Batches run sequentially; use [ProcessConcurrent]
// to process batches in parallel with a concurrency limit.
//
// [Chunk] splits a slice into sub-slices of at most the given size.
package batch
