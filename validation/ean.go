package validation

// EAN8 validates that value is a valid EAN-8 barcode.
func EAN8(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "EAN-8 is required"}
	}
	if len(value) != 8 {
		return &ValidationError{Field: field, Message: "EAN-8 must have 8 digits"}
	}
	if !eanCheck(value) {
		return &ValidationError{Field: field, Message: "invalid EAN-8 check digit"}
	}
	return nil
}

// EAN13 validates that value is a valid EAN-13 barcode.
func EAN13(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "EAN-13 is required"}
	}
	if len(value) != 13 {
		return &ValidationError{Field: field, Message: "EAN-13 must have 13 digits"}
	}
	if !eanCheck(value) {
		return &ValidationError{Field: field, Message: "invalid EAN-13 check digit"}
	}
	return nil
}

// UPC validates that value is a valid UPC-A code (12-digit EAN subset).
func UPC(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "UPC is required"}
	}
	if len(value) != 12 {
		return &ValidationError{Field: field, Message: "UPC-A must have 12 digits"}
	}
	if !eanCheck(value) {
		return &ValidationError{Field: field, Message: "invalid UPC check digit"}
	}
	return nil
}

// eanCheck validates the EAN/UPC check digit.
// Works for EAN-8, EAN-13, and UPC-A (same algorithm).
func eanCheck(s string) bool {
	n := len(s)
	sum := 0
	for i, c := range s {
		if c < '0' || c > '9' {
			return false
		}
		d := int(c - '0')
		// Digits at odd positions from the right get weight 3.
		if (n-1-i)%2 == 1 {
			d *= 3
		}
		sum += d
	}
	return sum%10 == 0
}
