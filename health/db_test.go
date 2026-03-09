package health

import (
	"context"
	"errors"
	"testing"
)

type mockPool struct {
	err error
}

func (m *mockPool) Ping(ctx context.Context) error { return m.err }

func TestDBChecker_Healthy(t *testing.T) {
	c := DBChecker(&mockPool{})
	if c.Name() != "db" {
		t.Errorf("expected name db, got %s", c.Name())
	}
	if err := c.Check(context.Background()); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestDBChecker_Unhealthy(t *testing.T) {
	c := DBChecker(&mockPool{err: errors.New("connection refused")})
	err := c.Check(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "connection refused" {
		t.Errorf("expected 'connection refused', got %s", err.Error())
	}
}
