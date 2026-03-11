package health

import (
	"context"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestMemoryChecker_Pass(t *testing.T) {
	// 1 GB threshold - should always pass in tests
	c := MemoryChecker(1 << 30)
	testkit.AssertEqual(t, c.Name(), "memory")
	testkit.AssertNoError(t, c.Check(context.Background()))
}

func TestMemoryChecker_Fail(t *testing.T) {
	// 1 byte threshold - should always fail
	c := MemoryChecker(1)
	err := c.Check(context.Background())
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "exceeds threshold")
}

func TestGoroutineChecker_Pass(t *testing.T) {
	// 100000 goroutine threshold - should pass in tests
	c := GoroutineChecker(100000)
	testkit.AssertEqual(t, c.Name(), "goroutines")
	testkit.AssertNoError(t, c.Check(context.Background()))
}

func TestGoroutineChecker_Fail(t *testing.T) {
	// 1 goroutine threshold - should fail (test itself uses more)
	c := GoroutineChecker(1)
	err := c.Check(context.Background())
	testkit.AssertError(t, err)
	testkit.AssertContains(t, err.Error(), "exceeds threshold")
}
