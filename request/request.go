// Package request provides HTTP request parsing helpers shared across microservices.
package request

import (
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
func ParsePagination(q url.Values, defaultLimit, maxLimit int) Pagination {
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
