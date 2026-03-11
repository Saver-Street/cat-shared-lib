package validation

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"
)

// ibanPattern matches a basic IBAN structure: 2 letters, 2 digits, then 11-30
// alphanumeric characters.
var ibanPattern = regexp.MustCompile(`^[A-Z]{2}\d{2}[A-Z0-9]{11,30}$`)

// ibanLengths maps country codes to their expected IBAN length per the IBAN
// registry. Only a subset of common countries is included; unknown countries
// fall back to the general pattern check and mod-97 validation.
var ibanLengths = map[string]int{
	"AL": 28, "AD": 24, "AT": 20, "AZ": 28, "BH": 22,
	"BY": 28, "BE": 16, "BA": 20, "BR": 29, "BG": 22,
	"CR": 22, "HR": 21, "CY": 28, "CZ": 24, "DK": 18,
	"DO": 28, "TL": 23, "EG": 29, "SV": 28, "EE": 20,
	"FO": 18, "FI": 18, "FR": 27, "GE": 22, "DE": 22,
	"GI": 23, "GR": 27, "GL": 18, "GT": 28, "HU": 28,
	"IS": 26, "IQ": 23, "IE": 22, "IL": 23, "IT": 27,
	"JO": 30, "KZ": 20, "XK": 20, "KW": 30, "LV": 21,
	"LB": 28, "LI": 21, "LT": 20, "LU": 20, "MT": 31,
	"MR": 27, "MU": 30, "MC": 27, "MD": 24, "ME": 22,
	"NL": 18, "MK": 19, "NO": 15, "PK": 24, "PS": 29,
	"PL": 28, "PT": 25, "QA": 29, "RO": 24, "LC": 32,
	"SM": 27, "ST": 25, "SA": 24, "RS": 22, "SC": 31,
	"SK": 24, "SI": 19, "ES": 24, "SE": 24, "CH": 21,
	"TN": 24, "TR": 26, "UA": 29, "AE": 23, "GB": 22,
	"VA": 22, "VG": 24,
}

// IBAN validates that value is a valid International Bank Account Number.
// It checks the format, country-specific length (when known), and the
// mod-97 check digits per ISO 13616.
func IBAN(field, value string) error {
	v := strings.ToUpper(strings.ReplaceAll(value, " ", ""))
	if !ibanPattern.MatchString(v) {
		return fmt.Errorf("%s must be a valid IBAN", field)
	}
	cc := v[:2]
	if expected, ok := ibanLengths[cc]; ok && len(v) != expected {
		return fmt.Errorf("%s must be a valid IBAN", field)
	}
	if !ibanMod97(v) {
		return fmt.Errorf("%s must be a valid IBAN", field)
	}
	return nil
}

// ibanMod97 rearranges the IBAN (move first 4 chars to end), converts letters
// to digits (A=10..Z=35), and checks that the resulting number mod 97 equals 1.
func ibanMod97(iban string) bool {
	rearranged := iban[4:] + iban[:4]
	var digits strings.Builder
	for _, ch := range rearranged {
		if ch >= 'A' && ch <= 'Z' {
			digits.WriteString(fmt.Sprintf("%d", ch-'A'+10))
		} else {
			digits.WriteByte(byte(ch))
		}
	}
	n := new(big.Int)
	if _, ok := n.SetString(digits.String(), 10); !ok {
		return false
	}
	mod := new(big.Int).Mod(n, big.NewInt(97))
	return mod.Int64() == 1
}
