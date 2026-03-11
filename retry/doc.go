// Package retry provides exponential backoff with jitter for retrying fallible
// operations.
//
// Call [Do] with a [Config] and a function to execute.  The function is
// retried up to [Config.MaxAttempts] times, with delays growing exponentially
// from [Config.InitialDelay] up to [Config.MaxDelay], multiplied by
// [Config.Multiplier] each attempt and randomized by [Config.JitterFraction].
//
// An optional [Config.RetryIf] predicate controls which errors trigger a
// retry; when nil every non-nil error is retried.  The operation respects
// context cancellation — if the context is done between attempts, [Do]
// returns the context error immediately.
package retry
