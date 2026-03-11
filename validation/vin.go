package validation

import "strings"

// vinWeights are the positional weight factors for VIN check digit (position 1–17).
var vinWeights = [17]int{8, 7, 6, 5, 4, 3, 2, 10, 0, 9, 8, 7, 6, 5, 4, 3, 2}

// vinTransliteration maps VIN characters to their numeric value.
var vinTransliteration = [256]int{}

func init() {
	values := map[byte]int{
		'A': 1, 'B': 2, 'C': 3, 'D': 4, 'E': 5, 'F': 6, 'G': 7, 'H': 8,
		'J': 1, 'K': 2, 'L': 3, 'M': 4, 'N': 5, 'P': 7, 'R': 9,
		'S': 2, 'T': 3, 'U': 4, 'V': 5, 'W': 6, 'X': 7, 'Y': 8, 'Z': 9,
	}
	for c := byte('0'); c <= '9'; c++ {
		vinTransliteration[c] = int(c - '0')
	}
	for c, v := range values {
		vinTransliteration[c] = v
	}
}

// VIN validates that value is a valid 17-character Vehicle Identification
// Number (ISO 3779) with a correct check digit (position 9).
// Letters I, O, and Q are not permitted.
func VIN(field, value string) *ValidationError {
	v := strings.ToUpper(strings.TrimSpace(value))
	if len(v) != 17 {
		return &ValidationError{Field: field, Message: "must be a valid 17-character VIN"}
	}
	for _, c := range v {
		if c == 'I' || c == 'O' || c == 'Q' {
			return &ValidationError{Field: field, Message: "must be a valid 17-character VIN"}
		}
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') {
			return &ValidationError{Field: field, Message: "must be a valid 17-character VIN"}
		}
	}
	sum := 0
	for i := range 17 {
		sum += vinTransliteration[v[i]] * vinWeights[i]
	}
	remainder := sum % 11
	var expected byte
	if remainder == 10 {
		expected = 'X'
	} else {
		expected = byte('0' + remainder)
	}
	if v[8] != expected {
		return &ValidationError{Field: field, Message: "must be a valid 17-character VIN"}
	}
	return nil
}
