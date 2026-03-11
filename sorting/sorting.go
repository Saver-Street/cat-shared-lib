// Package sorting provides helpers for parsing, validating, and applying
// sort parameters in list endpoints. It supports multi-field sorting with
// safe SQL ORDER BY clause generation.
package sorting

import (
	"fmt"
	"net/url"
	"strings"
)

// Direction represents a sort direction.
type Direction string

const (
	Asc  Direction = "asc"
	Desc Direction = "desc"
)

// Field represents a single sort field with direction.
type Field struct {
	Name      string    `json:"name"`
	Direction Direction `json:"direction"`
}

// String returns the field as "name asc" or "name desc".
func (f Field) String() string {
	return f.Name + " " + string(f.Direction)
}

// Params holds validated sort parameters.
type Params struct {
	Fields []Field
}

// OrderByClause returns a safe SQL ORDER BY clause (without the ORDER BY keyword).
// Column names are validated against the allowed set by Parse, so they are safe
// for interpolation. Returns "" if no fields are set.
func (p Params) OrderByClause() string {
	if len(p.Fields) == 0 {
		return ""
	}
	parts := make([]string, len(p.Fields))
	for i, f := range p.Fields {
		parts[i] = f.String()
	}
	return strings.Join(parts, ", ")
}

// HasField reports whether the params contain a field with the given name.
func (p Params) HasField(name string) bool {
	for _, f := range p.Fields {
		if f.Name == name {
			return true
		}
	}
	return false
}

// Config configures how sort parameters are parsed.
type Config struct {
	// Allowed is the set of permitted column names.
	Allowed []string
	// DefaultField is used when no sort parameter is provided.
	DefaultField string
	// DefaultDirection is used when no order parameter is provided.
	DefaultDirection Direction
	// MaxFields limits how many sort fields are accepted. 0 means 1.
	MaxFields int
}

func (c *Config) defaults() {
	if c.DefaultDirection == "" {
		c.DefaultDirection = Asc
	}
	if c.MaxFields <= 0 {
		c.MaxFields = 1
	}
}

func (c *Config) isAllowed(name string) bool {
	for _, a := range c.Allowed {
		if strings.EqualFold(a, name) {
			return true
		}
	}
	return false
}

func (c *Config) canonicalName(name string) string {
	for _, a := range c.Allowed {
		if strings.EqualFold(a, name) {
			return a
		}
	}
	return name
}

// Parse extracts sort parameters from URL query values.
//
// Supported query formats:
//   - Single: ?sort=name&order=asc
//   - Multi:  ?sort=name:asc,created_at:desc
//
// Fields not in Config.Allowed are silently dropped. If no valid fields
// remain, the default field and direction are used.
func Parse(q url.Values, cfg Config) Params {
	cfg.defaults()

	raw := strings.TrimSpace(q.Get("sort"))
	if raw == "" {
		return defaultParams(cfg)
	}

	var fields []Field

	// Check for multi-field format: "field:dir,field:dir"
	if strings.Contains(raw, ",") || strings.Contains(raw, ":") {
		parts := strings.Split(raw, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			f := parseFieldSpec(part, cfg)
			if f != nil && len(fields) < cfg.MaxFields {
				fields = append(fields, *f)
			}
		}
	} else if cfg.isAllowed(raw) {
		// Single field mode: sort=name&order=asc
		dir := parseDirection(q.Get("order"), cfg.DefaultDirection)
		fields = append(fields, Field{
			Name:      cfg.canonicalName(raw),
			Direction: dir,
		})
	}

	if len(fields) == 0 {
		return defaultParams(cfg)
	}
	return Params{Fields: fields}
}

func parseFieldSpec(spec string, cfg Config) *Field {
	parts := strings.SplitN(spec, ":", 2)
	name := strings.TrimSpace(parts[0])
	if !cfg.isAllowed(name) {
		return nil
	}
	dir := cfg.DefaultDirection
	if len(parts) == 2 {
		dir = parseDirection(parts[1], cfg.DefaultDirection)
	}
	return &Field{Name: cfg.canonicalName(name), Direction: dir}
}

func parseDirection(s string, fallback Direction) Direction {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "asc":
		return Asc
	case "desc":
		return Desc
	default:
		return fallback
	}
}

func defaultParams(cfg Config) Params {
	if cfg.DefaultField == "" {
		return Params{}
	}
	return Params{
		Fields: []Field{{Name: cfg.DefaultField, Direction: cfg.DefaultDirection}},
	}
}

// OrderBySQL returns a complete "ORDER BY ..." clause ready for appending to
// a query. Returns "" if params has no fields. The column names come from the
// validated allowed set, so they are safe for interpolation.
func OrderBySQL(p Params) string {
	clause := p.OrderByClause()
	if clause == "" {
		return ""
	}
	return fmt.Sprintf("ORDER BY %s", clause)
}
