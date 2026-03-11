// Package shutdown provides graceful shutdown utilities for HTTP servers with
// OS signal handling and in-flight connection draining.
//
// [ListenAndServe] wraps an *http.Server with SIGINT/SIGTERM handling and a
// configurable shutdown timeout from [Config].  [WaitForSignal] returns a
// context that cancels on the first received signal, useful for non-HTTP
// workloads.
//
// [Drainer] tracks in-flight requests via [Drainer.Add] and [Drainer.Done].
// Wrap an http.Handler with [Drainer.Middleware] to count active connections
// automatically, then call [Drainer.Wait] or [Drainer.WaitWithContext] during
// shutdown to block until all requests complete.  [Drainer.Active] and
// [Drainer.Completed] provide real-time observability into drain progress.
package shutdown
