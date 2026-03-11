package health

import (
	"context"
	"errors"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

type mockPool struct {
	err error
}

func (m *mockPool) Ping(ctx context.Context) error { return m.err }

func TestDBChecker_Healthy(t *testing.T) {
	c := DBChecker(&mockPool{})
	testkit.AssertEqual(t, c.Name(), "db")
	testkit.AssertNoError(t, c.Check(context.Background()))
}

func TestDBChecker_Unhealthy(t *testing.T) {
	c := DBChecker(&mockPool{err: errors.New("connection refused")})
	err := c.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	testkit.AssertEqual(t, err.Error(), "connection refused")
}

func BenchmarkDBChecker(b *testing.B) {
	c := DBChecker(&mockPool{})
	ctx := context.Background()
	for b.Loop() {
		c.Check(ctx)
	}
}
