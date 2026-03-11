package middleware

import "net/http"

// Chain composes middleware functions so that the first argument is the
// outermost wrapper.  The returned function wraps a handler with all
// provided middleware applied in order:
//
//	handler := middleware.Chain(logging, recovery, auth)(api)
//	// equivalent to: logging(recovery(auth(api)))
func Chain(mw ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			final = mw[i](final)
		}
		return final
	}
}
