package health

import (
	"context"
	"fmt"
	"runtime"
)

// MemoryChecker returns a Checker that fails when the process heap allocation
// exceeds maxBytes. Useful for detecting memory leaks in long-running services.
func MemoryChecker(maxBytes uint64) Checker {
	return NewChecker("memory", func(_ context.Context) error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.HeapAlloc > maxBytes {
			return fmt.Errorf("heap allocation %d bytes exceeds threshold %d bytes", m.HeapAlloc, maxBytes)
		}
		return nil
	})
}

// GoroutineChecker returns a Checker that fails when the number of goroutines
// exceeds maxGoroutines. Useful for detecting goroutine leaks.
func GoroutineChecker(maxGoroutines int) Checker {
	return NewChecker("goroutines", func(_ context.Context) error {
		n := runtime.NumGoroutine()
		if n > maxGoroutines {
			return fmt.Errorf("goroutine count %d exceeds threshold %d", n, maxGoroutines)
		}
		return nil
	})
}
