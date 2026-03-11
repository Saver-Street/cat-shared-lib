package validation

import (
	"time"
)

// Timezone validates that value is a valid IANA timezone name (e.g.
// "America/New_York", "UTC", "Europe/London").
func Timezone(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "timezone is required"}
	}
	_, err := time.LoadLocation(value)
	if err != nil {
		return &ValidationError{Field: field, Message: "invalid timezone"}
	}
	return nil
}

// TimezoneOffset validates that value is a valid fixed-offset timezone
// string in the format "+07:00", "-05:00", or "Z".
func TimezoneOffset(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "timezone offset is required"}
	}
	if value == "Z" {
		return nil
	}
	if len(value) != 6 {
		return &ValidationError{Field: field, Message: "invalid timezone offset format (expected ±HH:MM)"}
	}
	if value[0] != '+' && value[0] != '-' {
		return &ValidationError{Field: field, Message: "timezone offset must start with + or -"}
	}
	if value[3] != ':' {
		return &ValidationError{Field: field, Message: "invalid timezone offset format (expected ±HH:MM)"}
	}
	for _, i := range []int{1, 2, 4, 5} {
		if value[i] < '0' || value[i] > '9' {
			return &ValidationError{Field: field, Message: "timezone offset contains non-digit characters"}
		}
	}
	h := (int(value[1]-'0'))*10 + int(value[2]-'0')
	m := (int(value[4]-'0'))*10 + int(value[5]-'0')
	if h > 14 || m > 59 {
		return &ValidationError{Field: field, Message: "timezone offset out of range"}
	}
	if h == 14 && m > 0 {
		return &ValidationError{Field: field, Message: "timezone offset out of range"}
	}
	return nil
}
