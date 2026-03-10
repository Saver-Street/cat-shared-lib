// Package circuitbreaker implements the circuit breaker pattern for protecting
// external service calls. It tracks failures and temporarily halts requests
// to unhealthy services, allowing them time to recover.
package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// State represents the circuit breaker state.
type State int

const (
	// StateClosed allows requests through; failures are counted.
	StateClosed State = iota
	// StateOpen rejects all requests immediately.
	StateOpen
	// StateHalfOpen allows a limited number of probe requests.
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrOpenCircuit is returned when a call is rejected because the circuit is open.
var ErrOpenCircuit = errors.New("circuit breaker is open")

// Config configures a circuit breaker.
type Config struct {
	// MaxFailures is the number of consecutive failures before opening. Default: 5.
	MaxFailures int
	// Timeout is how long the circuit stays open before transitioning to half-open.
	// Default: 30 seconds.
	Timeout time.Duration
	// MaxHalfOpenRequests is the number of probe requests allowed in half-open state.
	// Default: 1.
	MaxHalfOpenRequests int
	// OnStateChange is called when the circuit breaker transitions between states.
	OnStateChange func(from, to State)
}

func (c *Config) defaults() {
	if c.MaxFailures <= 0 {
		c.MaxFailures = 5
	}
	if c.Timeout <= 0 {
		c.Timeout = 30 * time.Second
	}
	if c.MaxHalfOpenRequests <= 0 {
		c.MaxHalfOpenRequests = 1
	}
}

// Breaker is a circuit breaker for external service calls.
type Breaker struct {
	config           Config
	mu               sync.Mutex
	state            State
	failures         int
	successes        int
	halfOpenRequests int
	openedAt         time.Time
	now              func() time.Time
}

// New creates a new circuit breaker with the given configuration.
func New(cfg Config) *Breaker {
	cfg.defaults()
	return &Breaker{
		config: cfg,
		state:  StateClosed,
		now:    time.Now,
	}
}

// State returns the current state of the circuit breaker.
func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	// Check if an open circuit should transition to half-open.
	if b.state == StateOpen && b.now().Sub(b.openedAt) >= b.config.Timeout {
		b.transition(StateHalfOpen)
	}
	return b.state
}

// Do executes the given function if the circuit breaker allows it.
// It returns ErrOpenCircuit if the circuit is open.
func (b *Breaker) Do(fn func() error) error {
	b.mu.Lock()

	// Check for timeout-based transition from open to half-open.
	if b.state == StateOpen && b.now().Sub(b.openedAt) >= b.config.Timeout {
		b.transition(StateHalfOpen)
	}

	switch b.state {
	case StateOpen:
		b.mu.Unlock()
		return ErrOpenCircuit
	case StateHalfOpen:
		if b.halfOpenRequests >= b.config.MaxHalfOpenRequests {
			b.mu.Unlock()
			return ErrOpenCircuit
		}
		b.halfOpenRequests++
	}

	b.mu.Unlock()

	err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.recordFailure()
		return err
	}

	b.recordSuccess()
	return nil
}

func (b *Breaker) recordFailure() {
	b.failures++
	b.successes = 0
	if b.state == StateHalfOpen || b.failures >= b.config.MaxFailures {
		b.transition(StateOpen)
	}
}

func (b *Breaker) recordSuccess() {
	b.successes++
	b.failures = 0
	if b.state == StateHalfOpen {
		b.transition(StateClosed)
	}
}

func (b *Breaker) transition(to State) {
	from := b.state
	b.state = to
	b.failures = 0
	b.successes = 0
	b.halfOpenRequests = 0
	if to == StateOpen {
		b.openedAt = b.now()
	}
	if b.config.OnStateChange != nil {
		b.config.OnStateChange(from, to)
	}
}

// Reset resets the circuit breaker to the closed state.
func (b *Breaker) Reset() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.transition(StateClosed)
}

// Counts returns the current failure and success counts.
func (b *Breaker) Counts() (failures, successes int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.failures, b.successes
}
