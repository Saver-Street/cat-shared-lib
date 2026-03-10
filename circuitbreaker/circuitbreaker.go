// Package circuitbreaker implements the circuit breaker pattern for protecting
// calls to external services. A circuit breaker monitors failures and, after a
// configurable threshold, opens the circuit to prevent further calls, giving
// the downstream service time to recover.
//
// States:
//   - Closed: requests flow normally; failures are counted.
//   - Open: requests are rejected immediately with [ErrCircuitOpen].
//   - HalfOpen: a limited number of probe requests are allowed through to
//     test whether the downstream has recovered.
//
// Usage:
//
//	cb := circuitbreaker.New("payment-api",
//	    circuitbreaker.WithFailureThreshold(5),
//	    circuitbreaker.WithResetTimeout(30 * time.Second),
//	)
//
//	err := cb.Execute(func() error {
//	    return callPaymentAPI()
//	})
package circuitbreaker

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// State represents the current state of the circuit breaker.
type State int

const (
	// StateClosed indicates the circuit is closed and requests flow normally.
	StateClosed State = iota
	// StateOpen indicates the circuit is open and requests are rejected.
	StateOpen
	// StateHalfOpen indicates the circuit is testing whether the downstream has recovered.
	StateHalfOpen
)

// String returns the human-readable name of the state.
func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return fmt.Sprintf("unknown(%d)", int(s))
	}
}

// Sentinel errors returned by the circuit breaker.
var (
	// ErrCircuitOpen is returned when the circuit is open and calls are rejected.
	ErrCircuitOpen = errors.New("circuitbreaker: circuit is open")
	// ErrTooManyRequests is returned when too many requests are attempted in the half-open state.
	ErrTooManyRequests = errors.New("circuitbreaker: too many requests in half-open state")
)

// Counts tracks the number of requests and their results in the current state window.
type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

func (c *Counts) onSuccess() {
	c.Requests++
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
}

func (c *Counts) onFailure() {
	c.Requests++
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
}

func (c *Counts) reset() {
	*c = Counts{}
}

// StateChangeFunc is a callback invoked when the circuit breaker transitions between states.
type StateChangeFunc func(name string, from, to State)

// ReadyToTripFunc determines whether the circuit should trip (open) based on the current counts.
type ReadyToTripFunc func(counts Counts) bool

// IsSuccessfulFunc classifies whether an error should be treated as a success.
// Returning true means the call is counted as successful even though it returned an error.
type IsSuccessfulFunc func(err error) bool

// Options configures the circuit breaker behaviour.
type Options struct {
	// FailureThreshold is the number of consecutive failures before tripping the circuit.
	// Default: 5.
	FailureThreshold uint32

	// SuccessThreshold is the number of consecutive successes in the half-open state
	// required to close the circuit. Default: 2.
	SuccessThreshold uint32

	// MaxHalfOpenRequests is the maximum number of concurrent requests allowed in the
	// half-open state. Default: 1.
	MaxHalfOpenRequests uint32

	// ResetTimeout is how long the circuit stays open before transitioning to half-open.
	// Default: 60s.
	ResetTimeout time.Duration

	// OnStateChange is called when the circuit breaker transitions between states.
	OnStateChange StateChangeFunc

	// ReadyToTrip overrides the default failure-threshold logic. If set,
	// FailureThreshold is ignored.
	ReadyToTrip ReadyToTripFunc

	// IsSuccessful classifies whether an error should be treated as a success.
	// If nil, any non-nil error is treated as a failure.
	IsSuccessful IsSuccessfulFunc
}

// Option applies a configuration to the circuit breaker.
type Option func(*Options)

// WithFailureThreshold sets the number of consecutive failures before tripping.
func WithFailureThreshold(n uint32) Option {
	return func(o *Options) { o.FailureThreshold = n }
}

// WithSuccessThreshold sets the number of consecutive successes needed to close.
func WithSuccessThreshold(n uint32) Option {
	return func(o *Options) { o.SuccessThreshold = n }
}

// WithMaxHalfOpenRequests sets the maximum concurrent requests in half-open state.
func WithMaxHalfOpenRequests(n uint32) Option {
	return func(o *Options) { o.MaxHalfOpenRequests = n }
}

// WithResetTimeout sets how long the circuit stays open.
func WithResetTimeout(d time.Duration) Option {
	return func(o *Options) { o.ResetTimeout = d }
}

// WithOnStateChange registers a callback for state transitions.
func WithOnStateChange(fn StateChangeFunc) Option {
	return func(o *Options) { o.OnStateChange = fn }
}

// WithReadyToTrip overrides the default tripping logic.
func WithReadyToTrip(fn ReadyToTripFunc) Option {
	return func(o *Options) { o.ReadyToTrip = fn }
}

// WithIsSuccessful sets a custom error classifier.
func WithIsSuccessful(fn IsSuccessfulFunc) Option {
	return func(o *Options) { o.IsSuccessful = fn }
}

// Breaker implements the circuit breaker pattern.
type Breaker struct {
	name    string
	opts    Options
	mu      sync.Mutex
	state   State
	counts  Counts
	openAt  time.Time // when the circuit transitioned to open
	nowFunc func() time.Time
}

// New creates a circuit breaker with the given name and options.
func New(name string, opts ...Option) *Breaker {
	o := Options{
		FailureThreshold:    5,
		SuccessThreshold:    2,
		MaxHalfOpenRequests: 1,
		ResetTimeout:        60 * time.Second,
	}
	for _, fn := range opts {
		fn(&o)
	}
	return &Breaker{
		name:    name,
		opts:    o,
		state:   StateClosed,
		nowFunc: time.Now,
	}
}

// Name returns the circuit breaker's name.
func (b *Breaker) Name() string { return b.name }

// State returns the current state of the circuit breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.nowFunc()
	return b.currentState(now)
}

// Counts returns a snapshot of the current counters.
func (b *Breaker) Counts() Counts {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.counts
}

// Execute runs the given function within the circuit breaker. If the circuit
// is open, it returns [ErrCircuitOpen] without calling fn. If the circuit is
// half-open and the maximum number of probe requests has been reached, it
// returns [ErrTooManyRequests].
func (b *Breaker) Execute(fn func() error) error {
	if err := b.beforeRequest(); err != nil {
		return err
	}

	// panics count as failures
	var panicVal any
	defer func() {
		if panicVal != nil {
			b.afterRequest(false)
			panic(panicVal)
		}
	}()

	err := func() error {
		defer func() {
			if r := recover(); r != nil {
				panicVal = r
			}
		}()
		return fn()
	}()

	if panicVal != nil {
		return nil // deferred panic handler above will re-panic
	}

	success := err == nil
	if !success && b.opts.IsSuccessful != nil {
		success = b.opts.IsSuccessful(err)
	}
	b.afterRequest(success)
	return err
}

// Reset manually resets the circuit breaker to the closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.setState(StateClosed)
}

// currentState returns the effective state, accounting for the reset timeout.
// Must be called with b.mu held.
func (b *Breaker) currentState(now time.Time) State {
	if b.state == StateOpen && now.After(b.openAt.Add(b.opts.ResetTimeout)) {
		b.setState(StateHalfOpen)
	}
	return b.state
}

// setState transitions to the new state, resetting counters and invoking the callback.
// Must be called with b.mu held.
func (b *Breaker) setState(newState State) {
	if b.state == newState {
		return
	}
	prev := b.state
	b.state = newState
	b.counts.reset()

	if newState == StateOpen {
		b.openAt = b.nowFunc()
	}

	slog.Info("circuitbreaker: state change",
		"breaker", b.name,
		"from", prev.String(),
		"to", newState.String(),
	)

	if b.opts.OnStateChange != nil {
		b.opts.OnStateChange(b.name, prev, newState)
	}
}

// beforeRequest checks whether the request should be allowed. In the half-open
// state it pre-increments the request counter so that concurrent callers are
// properly bounded by MaxHalfOpenRequests.
func (b *Breaker) beforeRequest() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := b.nowFunc()
	state := b.currentState(now)

	switch state {
	case StateOpen:
		return ErrCircuitOpen
	case StateHalfOpen:
		if b.counts.Requests >= b.opts.MaxHalfOpenRequests {
			return ErrTooManyRequests
		}
		// pre-increment so concurrent callers see the in-flight request
		b.counts.Requests++
	}
	return nil
}

// afterRequest records the result of a request and transitions state if needed.
func (b *Breaker) afterRequest(success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if success {
		b.counts.onSuccess()
	} else {
		b.counts.onFailure()
	}

	switch b.state {
	case StateClosed:
		if b.shouldTrip() {
			b.setState(StateOpen)
		}
	case StateHalfOpen:
		if success && b.counts.ConsecutiveSuccesses >= b.opts.SuccessThreshold {
			b.setState(StateClosed)
		} else if !success {
			b.setState(StateOpen)
		}
	}
}

// shouldTrip returns true if the circuit should transition from closed to open.
// Must be called with b.mu held.
func (b *Breaker) shouldTrip() bool {
	if b.opts.ReadyToTrip != nil {
		return b.opts.ReadyToTrip(b.counts)
	}
	return b.counts.ConsecutiveFailures >= b.opts.FailureThreshold
}
