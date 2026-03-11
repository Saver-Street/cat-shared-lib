package validation

import (
	"strings"
	"unicode"
)

// CardNetwork represents a credit card network.
type CardNetwork string

const (
	// CardVisa identifies Visa cards.
	CardVisa CardNetwork = "visa"
	// CardMastercard identifies Mastercard cards.
	CardMastercard CardNetwork = "mastercard"
	// CardAmex identifies American Express cards.
	CardAmex CardNetwork = "amex"
	// CardDiscover identifies Discover cards.
	CardDiscover CardNetwork = "discover"
	// CardUnknown is returned when the card network cannot be determined.
	CardUnknown CardNetwork = "unknown"
)

// CreditCard validates that value is a valid credit card number using the
// Luhn algorithm.  Spaces and dashes are stripped before validation.
func CreditCard(field, value string) error {
	cleaned := stripCardChars(value)
	if len(cleaned) < 12 || len(cleaned) > 19 {
		return &ValidationError{Field: field, Message: "invalid credit card number length"}
	}
	for _, r := range cleaned {
		if !unicode.IsDigit(r) {
			return &ValidationError{Field: field, Message: "credit card number must contain only digits"}
		}
	}
	if !luhn(cleaned) {
		return &ValidationError{Field: field, Message: "invalid credit card number"}
	}
	return nil
}

// CreditCardNetwork validates a credit card number and checks that it belongs
// to one of the specified networks.
func CreditCardNetwork(field, value string, networks ...CardNetwork) error {
	if err := CreditCard(field, value); err != nil {
		return err
	}
	cleaned := stripCardChars(value)
	net := DetectCardNetwork(cleaned)
	for _, n := range networks {
		if net == n {
			return nil
		}
	}
	return &ValidationError{Field: field, Message: "card network not accepted"}
}

// DetectCardNetwork returns the card network based on the card number prefix
// and length. The input should contain only digits (no spaces or dashes).
func DetectCardNetwork(number string) CardNetwork {
	n := len(number)
	if n == 0 {
		return CardUnknown
	}
	switch {
	case strings.HasPrefix(number, "4") && (n == 13 || n == 16 || n == 19):
		return CardVisa
	case n == 16 && isMastercardPrefix(number):
		return CardMastercard
	case (strings.HasPrefix(number, "34") || strings.HasPrefix(number, "37")) && n == 15:
		return CardAmex
	case strings.HasPrefix(number, "6011") && n == 16:
		return CardDiscover
	case strings.HasPrefix(number, "65") && n == 16:
		return CardDiscover
	default:
		return CardUnknown
	}
}

func isMastercardPrefix(number string) bool {
	if len(number) < 2 {
		return false
	}
	prefix2 := (int(number[0]-'0'))*10 + int(number[1]-'0')
	if prefix2 >= 51 && prefix2 <= 55 {
		return true
	}
	if len(number) >= 4 {
		prefix4 := prefix2*100 + (int(number[2]-'0'))*10 + int(number[3]-'0')
		if prefix4 >= 2221 && prefix4 <= 2720 {
			return true
		}
	}
	return false
}

func stripCardChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if r != ' ' && r != '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// luhn implements the Luhn algorithm for checksum validation.
func luhn(number string) bool {
	sum := 0
	alt := false
	for i := len(number) - 1; i >= 0; i-- {
		d := int(number[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
