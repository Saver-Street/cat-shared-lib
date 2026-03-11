package validation

import "strings"

// ISBN10 validates that value is a valid ISBN-10.
// The value may contain hyphens, which are stripped before validation.
func ISBN10(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "ISBN-10 is required"}
	}
	digits := strings.ReplaceAll(value, "-", "")
	if len(digits) != 10 {
		return &ValidationError{Field: field, Message: "ISBN-10 must have 10 digits"}
	}
	sum := 0
	for i, c := range digits {
		var d int
		switch {
		case c >= '0' && c <= '9':
			d = int(c - '0')
		case (c == 'X' || c == 'x') && i == 9:
			d = 10
		default:
			return &ValidationError{Field: field, Message: "ISBN-10 contains invalid characters"}
		}
		sum += d * (10 - i)
	}
	if sum%11 != 0 {
		return &ValidationError{Field: field, Message: "invalid ISBN-10 check digit"}
	}
	return nil
}

// ISBN13 validates that value is a valid ISBN-13.
// The value may contain hyphens, which are stripped before validation.
func ISBN13(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "ISBN-13 is required"}
	}
	digits := strings.ReplaceAll(value, "-", "")
	if len(digits) != 13 {
		return &ValidationError{Field: field, Message: "ISBN-13 must have 13 digits"}
	}
	sum := 0
	for i, c := range digits {
		if c < '0' || c > '9' {
			return &ValidationError{Field: field, Message: "ISBN-13 contains invalid characters"}
		}
		d := int(c - '0')
		if i%2 == 1 {
			d *= 3
		}
		sum += d
	}
	if sum%10 != 0 {
		return &ValidationError{Field: field, Message: "invalid ISBN-13 check digit"}
	}
	return nil
}

// ISBN validates that value is either a valid ISBN-10 or ISBN-13.
func ISBN(field, value string) error {
	digits := strings.ReplaceAll(value, "-", "")
	switch len(digits) {
	case 10:
		return ISBN10(field, value)
	case 13:
		return ISBN13(field, value)
	default:
		return &ValidationError{Field: field, Message: "must be a valid ISBN-10 or ISBN-13"}
	}
}
