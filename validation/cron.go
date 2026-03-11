package validation

import (
	"strconv"
	"strings"
)

// CronExpression validates a standard 5-field cron expression
// (minute hour day-of-month month day-of-week). Each field may contain
// numbers, ranges (1-5), steps (*/2), lists (1,3,5), and the wildcard *.
func CronExpression(field, value string) error {
	fields := strings.Fields(value)
	if len(fields) != 5 {
		return &ValidationError{Field: field, Message: "must be a valid cron expression (5 fields)"}
	}
	type fieldDef struct {
		name string
		min  int
		max  int
	}
	defs := [5]fieldDef{
		{"minute", 0, 59},
		{"hour", 0, 23},
		{"day-of-month", 1, 31},
		{"month", 1, 12},
		{"day-of-week", 0, 6},
	}
	for i, f := range fields {
		if err := validateCronField(f, defs[i].min, defs[i].max); err != nil {
			return &ValidationError{
				Field:   field,
				Message: "invalid " + defs[i].name + ": " + err.Error(),
			}
		}
	}
	return nil
}

func validateCronField(s string, min, max int) error {
	for _, part := range strings.Split(s, ",") {
		if err := validateCronPart(part, min, max); err != nil {
			return err
		}
	}
	return nil
}

func validateCronPart(s string, min, max int) error {
	base, step, hasStep := strings.Cut(s, "/")
	if hasStep {
		sv, err := strconv.Atoi(step)
		if err != nil || sv < 1 {
			return &cronFieldError{what: "step value"}
		}
		_ = sv
	}

	if base == "*" {
		return nil
	}

	if lo, hi, ok := strings.Cut(base, "-"); ok {
		loV, err := strconv.Atoi(lo)
		if err != nil || loV < min || loV > max {
			return &cronFieldError{what: "range start"}
		}
		hiV, err := strconv.Atoi(hi)
		if err != nil || hiV < min || hiV > max {
			return &cronFieldError{what: "range end"}
		}
		if loV > hiV {
			return &cronFieldError{what: "range start > end"}
		}
		return nil
	}

	v, err := strconv.Atoi(base)
	if err != nil || v < min || v > max {
		return &cronFieldError{what: "value"}
	}
	return nil
}

type cronFieldError struct {
	what string
}

func (e *cronFieldError) Error() string {
	return "invalid " + e.what
}
