package circuitbreaker

import (
	"errors"
	"testing"
	"time"
)

func FuzzExecute(f *testing.F) {
	f.Add("breaker-1", true)
	f.Add("", false)
	f.Add("unicode-⚡", true)
	f.Add("very-long-name-for-a-circuit-breaker", false)
	f.Fuzz(func(t *testing.T, name string, succeed bool) {
		cb := New(name,
			WithFailureThreshold(3),
			WithSuccessThreshold(2),
			WithResetTimeout(time.Second),
		)

		fn := func() error {
			if !succeed {
				return errors.New("fail")
			}
			return nil
		}

		// Execute must not panic.
		_ = cb.Execute(fn)

		// Name must return what was given.
		if cb.Name() != name {
			t.Errorf("Name() = %q, want %q", cb.Name(), name)
		}

		// State and Counts must not panic.
		_ = cb.State()
		_ = cb.Counts()
	})
}

func FuzzStateString(f *testing.F) {
	f.Add(uint32(0))
	f.Add(uint32(1))
	f.Add(uint32(2))
	f.Add(uint32(99))
	f.Fuzz(func(t *testing.T, s uint32) {
		state := State(s)
		// String() must not panic on any value.
		result := state.String()
		if result == "" {
			t.Error("String() returned empty")
		}
	})
}

func FuzzExecuteTransitions(f *testing.F) {
	f.Add(uint8(5), uint8(3))
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(10), uint8(10))
	f.Fuzz(func(t *testing.T, failures, successes uint8) {
		cb := New("fuzz-transitions",
			WithFailureThreshold(3),
			WithSuccessThreshold(2),
			WithResetTimeout(time.Millisecond),
		)

		// Apply failures then successes; must not panic.
		for range failures {
			cb.Execute(func() error { return errors.New("err") })
		}
		for range successes {
			cb.Execute(func() error { return nil })
		}

		// State must be valid.
		s := cb.State()
		if s != StateClosed && s != StateOpen && s != StateHalfOpen {
			t.Errorf("unexpected state: %v", s)
		}
	})
}
