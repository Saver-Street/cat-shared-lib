package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

var errService = errors.New("service unavailable")

func TestNew_Defaults(t *testing.T) {
	cb := New(Config{})
	if cb.config.MaxFailures != 5 {
		t.Errorf("expected 5 max failures, got %d", cb.config.MaxFailures)
	}
	if cb.config.Timeout != 30*time.Second {
		t.Errorf("expected 30s timeout, got %v", cb.config.Timeout)
	}
	if cb.config.MaxHalfOpenRequests != 1 {
		t.Errorf("expected 1 max half-open request, got %d", cb.config.MaxHalfOpenRequests)
	}
	if cb.State() != StateClosed {
		t.Errorf("expected closed state, got %v", cb.State())
	}
}

func TestStateString(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("State(%d).String() = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestDo_SuccessInClosed(t *testing.T) {
	cb := New(Config{MaxFailures: 3})
	err := cb.Do(func() error { return nil })
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	f, s := cb.Counts()
	if f != 0 || s != 1 {
		t.Errorf("expected 0 failures 1 success, got %d/%d", f, s)
	}
}

func TestDo_FailureCountsUp(t *testing.T) {
	cb := New(Config{MaxFailures: 3})
	for i := 0; i < 2; i++ {
		cb.Do(func() error { return errService })
	}
	f, _ := cb.Counts()
	if f != 2 {
		t.Errorf("expected 2 failures, got %d", f)
	}
	if cb.State() != StateClosed {
		t.Error("should still be closed")
	}
}

func TestDo_OpensAfterMaxFailures(t *testing.T) {
	var transitions []string
	cb := New(Config{
		MaxFailures: 3,
		OnStateChange: func(from, to State) {
			transitions = append(transitions, from.String()+"->"+to.String())
		},
	})

	for i := 0; i < 3; i++ {
		cb.Do(func() error { return errService })
	}

	if cb.State() != StateOpen {
		t.Errorf("expected open state, got %v", cb.State())
	}
	if len(transitions) != 1 || transitions[0] != "closed->open" {
		t.Errorf("expected closed->open transition, got %v", transitions)
	}
}

func TestDo_RejectsWhenOpen(t *testing.T) {
	cb := New(Config{MaxFailures: 1})
	cb.Do(func() error { return errService })

	err := cb.Do(func() error { return nil })
	if !errors.Is(err, ErrOpenCircuit) {
		t.Errorf("expected ErrOpenCircuit, got %v", err)
	}
}

func TestDo_TransitionsToHalfOpen(t *testing.T) {
	now := time.Now()
	cb := New(Config{MaxFailures: 1, Timeout: time.Second})
	cb.now = func() time.Time { return now }

	cb.Do(func() error { return errService })
	if cb.State() != StateOpen {
		t.Fatal("should be open")
	}

	// Advance past timeout.
	now = now.Add(2 * time.Second)
	if cb.State() != StateHalfOpen {
		t.Errorf("expected half-open, got %v", cb.State())
	}
}

func TestDo_HalfOpenSuccess_Closes(t *testing.T) {
	now := time.Now()
	cb := New(Config{MaxFailures: 1, Timeout: time.Second, MaxHalfOpenRequests: 1})
	cb.now = func() time.Time { return now }

	cb.Do(func() error { return errService })

	now = now.Add(2 * time.Second)
	err := cb.Do(func() error { return nil })
	if err != nil {
		t.Errorf("expected nil, got %v", err)
	}
	if cb.State() != StateClosed {
		t.Errorf("expected closed after half-open success, got %v", cb.State())
	}
}

func TestDo_HalfOpenFailure_ReOpens(t *testing.T) {
	now := time.Now()
	cb := New(Config{MaxFailures: 1, Timeout: time.Second, MaxHalfOpenRequests: 1})
	cb.now = func() time.Time { return now }

	cb.Do(func() error { return errService })

	now = now.Add(2 * time.Second)
	cb.Do(func() error { return errService })
	if cb.State() != StateOpen {
		t.Errorf("expected re-opened, got %v", cb.State())
	}
}

func TestDo_HalfOpenExceedsMaxRequests(t *testing.T) {
	now := time.Now()
	cb := New(Config{MaxFailures: 1, Timeout: time.Second, MaxHalfOpenRequests: 1})
	cb.now = func() time.Time { return now }

	cb.Do(func() error { return errService })
	now = now.Add(2 * time.Second)

	// First half-open request should be allowed (will block).
	started := make(chan struct{})
	done := make(chan struct{})
	go func() {
		cb.mu.Lock()
		// Simulate transition to half-open manually.
		if cb.state == StateOpen {
			cb.transition(StateHalfOpen)
		}
		cb.halfOpenRequests = cb.config.MaxHalfOpenRequests
		cb.mu.Unlock()
		close(started)

		// Second request should be rejected.
		err := cb.Do(func() error { return nil })
		if !errors.Is(err, ErrOpenCircuit) {
			t.Errorf("expected ErrOpenCircuit for excess half-open request, got %v", err)
		}
		close(done)
	}()

	<-started
	<-done
}

func TestReset(t *testing.T) {
	cb := New(Config{MaxFailures: 1})
	cb.Do(func() error { return errService })
	if cb.State() != StateOpen {
		t.Fatal("should be open")
	}
	cb.Reset()
	if cb.State() != StateClosed {
		t.Errorf("expected closed after reset, got %v", cb.State())
	}
}

func TestDo_SuccessResetsFailureCount(t *testing.T) {
	cb := New(Config{MaxFailures: 3})
	cb.Do(func() error { return errService })
	cb.Do(func() error { return errService })
	cb.Do(func() error { return nil })

	f, s := cb.Counts()
	if f != 0 {
		t.Errorf("expected 0 failures after success, got %d", f)
	}
	if s != 1 {
		t.Errorf("expected 1 success, got %d", s)
	}
}

func TestConcurrent_Do(t *testing.T) {
	cb := New(Config{MaxFailures: 100, Timeout: time.Second})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			if n%2 == 0 {
				cb.Do(func() error { return nil })
			} else {
				cb.Do(func() error { return errService })
			}
		}(i)
	}
	wg.Wait()
}

func TestDo_ReturnsOriginalError(t *testing.T) {
	custom := errors.New("custom error")
	cb := New(Config{MaxFailures: 5})
	err := cb.Do(func() error { return custom })
	if !errors.Is(err, custom) {
		t.Errorf("expected original error, got %v", err)
	}
}

func BenchmarkDo_Closed(b *testing.B) {
	cb := New(Config{MaxFailures: 1000000})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Do(func() error { return nil })
	}
}
