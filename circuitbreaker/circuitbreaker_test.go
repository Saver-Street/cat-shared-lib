package circuitbreaker

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestNew_Defaults(t *testing.T) {
	cb := New("test")
	testkit.AssertEqual(t, cb.Name(), "test")
	testkit.AssertEqual(t, cb.State(), StateClosed)
	testkit.AssertEqual(t, cb.opts.FailureThreshold, uint32(5))
	testkit.AssertEqual(t, cb.opts.SuccessThreshold, uint32(2))
	testkit.AssertEqual(t, cb.opts.MaxHalfOpenRequests, uint32(1))
	testkit.AssertEqual(t, cb.opts.ResetTimeout, 60*time.Second)
}

func TestNew_WithOptions(t *testing.T) {
	cb := New("test",
		WithFailureThreshold(3),
		WithSuccessThreshold(1),
		WithMaxHalfOpenRequests(2),
		WithResetTimeout(10*time.Second),
	)
	testkit.AssertEqual(t, cb.opts.FailureThreshold, uint32(3))
	testkit.AssertEqual(t, cb.opts.SuccessThreshold, uint32(1))
	testkit.AssertEqual(t, cb.opts.MaxHalfOpenRequests, uint32(2))
	testkit.AssertEqual(t, cb.opts.ResetTimeout, 10*time.Second)
}

func TestState_String(t *testing.T) {
	tests := []struct {
		state State
		want  string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown(99)"},
	}
	for _, tt := range tests {
		testkit.AssertEqual(t, tt.state.String(), tt.want)
	}
}

func TestExecute_Success(t *testing.T) {
	cb := New("test")
	err := cb.Execute(func() error { return nil })
	if err != nil {
		t.Fatalf("Execute() = %v, want nil", err)
	}
	c := cb.Counts()
	testkit.AssertEqual(t, c.TotalSuccesses, uint32(1))
	testkit.AssertEqual(t, c.ConsecutiveSuccesses, uint32(1))
}

func TestExecute_Failure(t *testing.T) {
	testErr := errors.New("fail")
	cb := New("test")
	err := cb.Execute(func() error { return testErr })
	testkit.AssertErrorIs(t, err, testErr)
	c := cb.Counts()
	testkit.AssertEqual(t, c.TotalFailures, uint32(1))
	testkit.AssertEqual(t, c.ConsecutiveFailures, uint32(1))
}

func TestTrip_AfterConsecutiveFailures(t *testing.T) {
	cb := New("test", WithFailureThreshold(3))
	fail := func() error { return errors.New("fail") }

	for i := 0; i < 3; i++ {
		_ = cb.Execute(fail)
	}

	testkit.AssertEqual(t, cb.State(), StateOpen)

	// subsequent calls should be rejected
	err := cb.Execute(func() error { return nil })
	testkit.AssertErrorIs(t, err, ErrCircuitOpen)
}

func TestTrip_SuccessResetsConsecutiveFailures(t *testing.T) {
	cb := New("test", WithFailureThreshold(3))
	fail := func() error { return errors.New("fail") }
	ok := func() error { return nil }

	_ = cb.Execute(fail)
	_ = cb.Execute(fail)
	_ = cb.Execute(ok) // resets consecutive
	_ = cb.Execute(fail)
	_ = cb.Execute(fail)

	testkit.AssertEqual(t, cb.State(), StateClosed)
}

func TestHalfOpen_TransitionAfterTimeout(t *testing.T) {
	now := time.Now()
	cb := New("test", WithFailureThreshold(1), WithResetTimeout(100*time.Millisecond))
	cb.nowFunc = func() time.Time { return now }

	// trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })
	if cb.State() != StateOpen {
		t.Fatalf("State() = %v, want %v", cb.State(), StateOpen)
	}

	// advance time past reset timeout
	now = now.Add(200 * time.Millisecond)
	testkit.AssertEqual(t, cb.State(), StateHalfOpen)
}

func TestHalfOpen_SuccessCloses(t *testing.T) {
	now := time.Now()
	cb := New("test",
		WithFailureThreshold(1),
		WithSuccessThreshold(2),
		WithMaxHalfOpenRequests(3),
		WithResetTimeout(100*time.Millisecond),
	)
	cb.nowFunc = func() time.Time { return now }

	// trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })

	// advance time to half-open
	now = now.Add(200 * time.Millisecond)

	// two successes should close
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return nil })

	testkit.AssertEqual(t, cb.State(), StateClosed)
}

func TestHalfOpen_FailureReopens(t *testing.T) {
	now := time.Now()
	cb := New("test",
		WithFailureThreshold(1),
		WithResetTimeout(100*time.Millisecond),
	)
	cb.nowFunc = func() time.Time { return now }

	// trip the breaker
	_ = cb.Execute(func() error { return errors.New("fail") })

	// advance to half-open
	now = now.Add(200 * time.Millisecond)

	// failure in half-open re-opens
	_ = cb.Execute(func() error { return errors.New("still broken") })

	testkit.AssertEqual(t, cb.State(), StateOpen)
}

func TestHalfOpen_TooManyRequests(t *testing.T) {
	now := time.Now()
	cb := New("test",
		WithFailureThreshold(1),
		WithMaxHalfOpenRequests(1),
		WithResetTimeout(100*time.Millisecond),
	)
	cb.nowFunc = func() time.Time { return now }

	// trip
	_ = cb.Execute(func() error { return errors.New("fail") })

	// advance to half-open
	now = now.Add(200 * time.Millisecond)

	// first request allowed (increments count to 1)
	done := make(chan struct{})
	started := make(chan struct{})
	go func() {
		_ = cb.Execute(func() error {
			close(started)
			<-done
			return nil
		})
	}()
	<-started

	// second request should be rejected
	err := cb.Execute(func() error { return nil })
	testkit.AssertErrorIs(t, err, ErrTooManyRequests)
	close(done)
}

func TestOnStateChange_Callback(t *testing.T) {
	var transitions []struct{ from, to State }
	cb := New("test",
		WithFailureThreshold(1),
		WithOnStateChange(func(name string, from, to State) {
			transitions = append(transitions, struct{ from, to State }{from, to})
		}),
	)

	_ = cb.Execute(func() error { return errors.New("fail") })

	if len(transitions) != 1 {
		t.Fatalf("expected 1 transition, got %d", len(transitions))
	}
	testkit.AssertEqual(t, transitions[0].from, StateClosed)
	testkit.AssertEqual(t, transitions[0].to, StateOpen)
}

func TestReadyToTrip_Custom(t *testing.T) {
	cb := New("test",
		WithReadyToTrip(func(c Counts) bool {
			// trip after 50% failure rate with at least 4 requests
			return c.Requests >= 4 && c.TotalFailures*2 >= c.Requests
		}),
	)

	// 2 successes, 2 failures (50% failure rate at 4 requests)
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return errors.New("fail") })
	_ = cb.Execute(func() error { return nil })
	_ = cb.Execute(func() error { return errors.New("fail") })

	testkit.AssertEqual(t, cb.State(), StateOpen)
}

func TestIsSuccessful_Custom(t *testing.T) {
	var errExpected = errors.New("expected")
	cb := New("test",
		WithFailureThreshold(1),
		WithIsSuccessful(func(err error) bool {
			return errors.Is(err, errExpected)
		}),
	)

	// errExpected should be treated as success
	err := cb.Execute(func() error { return errExpected })
	testkit.AssertErrorIs(t, err, errExpected)
	testkit.AssertEqual(t, cb.State(), StateClosed)
	c := cb.Counts()
	testkit.AssertEqual(t, c.ConsecutiveSuccesses, uint32(1))
}

func TestReset(t *testing.T) {
	cb := New("test", WithFailureThreshold(1))
	_ = cb.Execute(func() error { return errors.New("fail") })

	if cb.State() != StateOpen {
		t.Fatalf("State() = %v, want %v", cb.State(), StateOpen)
	}

	cb.Reset()
	testkit.AssertEqual(t, cb.State(), StateClosed)
	c := cb.Counts()
	testkit.AssertEqual(t, c.Requests, uint32(0))
}

func TestExecute_Panic(t *testing.T) {
	cb := New("test", WithFailureThreshold(2))

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic to propagate")
		}

		// panic should count as failure
		c := cb.Counts()
		testkit.AssertEqual(t, c.TotalFailures, uint32(1))
	}()

	_ = cb.Execute(func() error {
		panic("boom")
	})
}

func TestConcurrentAccess(t *testing.T) {
	cb := New("test", WithFailureThreshold(100))
	var wg sync.WaitGroup
	const goroutines = 50
	const iterations = 100

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = cb.Execute(func() error {
					if j%3 == 0 {
						return errors.New("fail")
					}
					return nil
				})
			}
		}()
	}
	wg.Wait()

	c := cb.Counts()
	testkit.AssertEqual(t, c.Requests, uint32(goroutines*iterations))
}

func TestCounts_Reset(t *testing.T) {
	var c Counts
	c.onSuccess()
	c.onSuccess()
	c.onFailure()

	testkit.AssertEqual(t, c.Requests, uint32(3))

	c.reset()
	testkit.AssertEqual(t, c.Requests, uint32(0))
	testkit.AssertEqual(t, c.TotalSuccesses, uint32(0))
	testkit.AssertEqual(t, c.TotalFailures, uint32(0))
}

func TestFullLifecycle(t *testing.T) {
	now := time.Now()
	var transitions []string
	cb := New("lifecycle",
		WithFailureThreshold(2),
		WithSuccessThreshold(1),
		WithResetTimeout(50*time.Millisecond),
		WithOnStateChange(func(name string, from, to State) {
			transitions = append(transitions, from.String()+"→"+to.String())
		}),
	)
	cb.nowFunc = func() time.Time { return now }

	// 1. closed → open (2 failures)
	_ = cb.Execute(func() error { return errors.New("err") })
	_ = cb.Execute(func() error { return errors.New("err") })
	if cb.State() != StateOpen {
		t.Fatalf("step 1: State() = %v, want open", cb.State())
	}

	// 2. open → half-open (timeout)
	now = now.Add(100 * time.Millisecond)
	if cb.State() != StateHalfOpen {
		t.Fatalf("step 2: State() = %v, want half-open", cb.State())
	}

	// 3. half-open → closed (success)
	_ = cb.Execute(func() error { return nil })
	if cb.State() != StateClosed {
		t.Fatalf("step 3: State() = %v, want closed", cb.State())
	}

	expected := []string{"closed→open", "open→half-open", "half-open→closed"}
	if len(transitions) != len(expected) {
		t.Fatalf("transitions = %v, want %v", transitions, expected)
	}
	for i, tr := range transitions {
		testkit.AssertEqual(t, tr, expected[i])
	}
}

func TestSetState_SameState_Noop(t *testing.T) {
	cb := New("test")
	cb.mu.Lock()
	stateBefore := cb.state
	countsBefore := cb.counts
	cb.setState(StateClosed) // already closed → no-op
	stateAfter := cb.state
	countsAfter := cb.counts
	cb.mu.Unlock()

	testkit.AssertEqual(t, stateAfter, stateBefore)
	testkit.AssertEqual(t, countsAfter, countsBefore)
}
