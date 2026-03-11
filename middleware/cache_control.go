package middleware

import (
"fmt"
"net/http"
"time"
)

// CacheControl returns middleware that sets a Cache-Control header with the
// given max-age duration. Use for static assets, public API responses, or
// any endpoint whose response can be cached for a known duration.
//
// If public is true, the directive includes "public"; otherwise "private".
func CacheControl(maxAge time.Duration, public bool) func(http.Handler) http.Handler {
visibility := "private"
if public {
visibility = "public"
}
value := fmt.Sprintf("%s, max-age=%d", visibility, int(maxAge.Seconds()))
return func(next http.Handler) http.Handler {
return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Cache-Control", value)
next.ServeHTTP(w, r)
})
}
}
