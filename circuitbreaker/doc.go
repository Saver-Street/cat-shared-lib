// Package circuitbreaker implements the circuit breaker pattern for protecting
// calls to external services from cascading failures.
//
// A [Breaker] transitions between three states: [StateClosed] (normal
// operation), [StateOpen] (calls rejected with [ErrCircuitOpen]), and
// [StateHalfOpen] (limited probe requests allowed).  Create one with [New]
// and functional [Option] values such as [WithFailureThreshold],
// [WithSuccessThreshold], [WithResetTimeout], and [WithMaxHalfOpenRequests].
//
// Call [Breaker.Execute] to run a function through the breaker.  Failures
// increment internal [Counts]; when the threshold is reached the circuit
// opens.  After the reset timeout elapses, probe requests are allowed and
// consecutive successes close the circuit.  [WithReadyToTrip] and
// [WithIsSuccessful] allow custom trip and error-classification logic.
// [WithOnStateChange] registers a callback for state transitions.
package circuitbreaker
