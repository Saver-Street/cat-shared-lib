// Package sorting provides helpers for parsing, validating, and applying
// sort parameters in list endpoints.
//
// It supports single-field ("?sort=name&order=asc") and multi-field
// ("?sort=name:asc,created_at:desc") query formats.  Parsed fields are
// validated against an allowed set, making the resulting SQL ORDER BY
// clause safe for interpolation.
//
// # Quick Start
//
//	cfg := sorting.Config{
//	    Allowed:          []string{"name", "created_at"},
//	    DefaultField:     "created_at",
//	    DefaultDirection: sorting.Desc,
//	    MaxFields:        3,
//	}
//	params := sorting.Parse(r.URL.Query(), cfg)
//	clause := sorting.OrderBySQL(params)  // "ORDER BY created_at desc"
package sorting
