package validation

import "strings"

// ISSN validates that value is a valid ISSN (International Standard Serial
// Number) in the format NNNN-NNNC where C is a check digit (0-9 or X).
// The check digit is verified using the ISSN MOD-11 algorithm.
func ISSN(field, value string) *ValidationError {
	v := strings.TrimSpace(value)
	msg := "must be a valid ISSN (e.g., 0378-5955)"
	if len(v) != 9 || v[4] != '-' {
		return &ValidationError{Field: field, Message: msg}
	}
	sum := 0
	weight := 8
	for i := range 9 {
		if i == 4 {
			continue
		}
		c := v[i]
		if i == 8 && (c == 'X' || c == 'x') {
			sum += 10 * weight
		} else if c >= '0' && c <= '9' {
			sum += int(c-'0') * weight
		} else {
			return &ValidationError{Field: field, Message: msg}
		}
		weight--
	}
	if sum%11 != 0 {
		return &ValidationError{Field: field, Message: msg}
	}
	return nil
}
