// Package sorting provides helpers for parsing, validating, and applying
// sort parameters in list endpoints. It supports multi-field sorting with
// safe SQL ORDER BY clause generation.
//
// # Parsing Sort Parameters
//
// Use [Parse] to extract sort parameters from URL query strings:
//
//	cfg := sorting.Config{
//	    Allowed:          []string{"name", "created_at"},
//	    DefaultField:     "created_at",
//	    DefaultDirection: Desc,
//	}
//	params := sorting.Parse(r.URL.Query(), cfg)
//
// # SQL Generation
//
// Use [OrderBySQL] or [Params.OrderByClause] to produce safe ORDER BY
// clauses for SQL queries:
//
//	clause := sorting.OrderBySQL(params) // "ORDER BY name ASC"
//
// Fields are validated against [Config.Allowed] to prevent SQL
// injection. Unknown fields are silently dropped.
package sorting
