// Package request provides HTTP request parsing helpers shared across microservices.
package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Pagination holds parsed page/limit/offset values from query parameters.
type Pagination struct {
	Page   int
	Limit  int
	Offset int
}

// ParsePagination extracts page and limit from URL query parameters with defaults and bounds.
// Zero or negative defaultLimit defaults to 20; zero or negative maxLimit defaults to 100.
func ParsePagination(q url.Values, defaultLimit, maxLimit int) Pagination {
	if defaultLimit <= 0 {
		defaultLimit = 20
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}
	page := 1
	if p := q.Get("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := defaultLimit
	if l := q.Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return Pagination{
		Page:   page,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
}

// URLParamFunc is a function that extracts a named URL parameter from a request.
// This allows callers to plug in any router (chi, gorilla/mux, etc.)
type URLParamFunc func(r *http.Request, key string) string

// RequireURLParam extracts a URL parameter using the provided paramFn.
// Returns an error if the parameter is empty or missing.
func RequireURLParam(r *http.Request, key string, paramFn URLParamFunc) (string, error) {
	val := paramFn(r, key)
	if val == "" {
		return "", fmt.Errorf("missing required URL parameter: %s", key)
	}
	return val, nil
}

// RequireURLParamInt extracts a URL parameter as an int64.
// Returns an error if the parameter is empty, missing, or not a valid integer.
func RequireURLParamInt(r *http.Request, key string, paramFn URLParamFunc) (int64, error) {
	val, err := RequireURLParam(r, key, paramFn)
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("URL parameter %q must be an integer, got %q", key, val)
	}
	if n <= 0 {
		return 0, fmt.Errorf("URL parameter %q must be positive, got %d", key, n)
	}
	return n, nil
}
