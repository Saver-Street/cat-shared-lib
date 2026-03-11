package validation

import (
	"fmt"
	"time"
)

// Date validates that value parses with the given layout (e.g., time.DateOnly).
// Returns a *ValidationError on failure.
func Date(field, value, layout string) error {
	if _, err := time.Parse(layout, value); err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid date (%s)", field, layout),
		}
	}
	return nil
}

// DateBefore validates that value is strictly before the boundary time.
func DateBefore(field, value, layout string, before time.Time) error {
	t, err := time.Parse(layout, value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid date (%s)", field, layout),
		}
	}
	if !t.Before(before) {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be before %s", field, before.Format(layout)),
		}
	}
	return nil
}

// DateAfter validates that value is strictly after the boundary time.
func DateAfter(field, value, layout string, after time.Time) error {
	t, err := time.Parse(layout, value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid date (%s)", field, layout),
		}
	}
	if !t.After(after) {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be after %s", field, after.Format(layout)),
		}
	}
	return nil
}

// DateRange validates that value falls within [earliest, latest] inclusive.
func DateRange(field, value, layout string, earliest, latest time.Time) error {
	t, err := time.Parse(layout, value)
	if err != nil {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be a valid date (%s)", field, layout),
		}
	}
	if t.Before(earliest) || t.After(latest) {
		return &ValidationError{
			Field: field,
			Message: fmt.Sprintf("%s must be between %s and %s",
				field, earliest.Format(layout), latest.Format(layout)),
		}
	}
	return nil
}

// FutureDate validates that value parses to a time strictly after now.
func FutureDate(field, value, layout string) error {
	return DateAfter(field, value, layout, time.Now())
}

// PastDate validates that value parses to a time strictly before now.
func PastDate(field, value, layout string) error {
	return DateBefore(field, value, layout, time.Now())
}
