package validation

import (
	"regexp"
	"strings"
)

// bcp47Re matches BCP 47 language tags like "en", "en-US", "zh-Hant-TW".
var bcp47Re = regexp.MustCompile(
	`^[a-zA-Z]{2,3}` + // primary language subtag
		`(-[a-zA-Z]{4})?` + // optional script subtag
		`(-[a-zA-Z]{2}|-[0-9]{3})?` + // optional region subtag
		`(-([a-zA-Z0-9]{5,8}|[0-9][a-zA-Z0-9]{3}))*$`, // optional variant subtags
)

// Locale validates that value is a well-formed BCP 47 language tag
// (e.g. "en", "en-US", "zh-Hant-TW"). It checks structure, not
// whether the tag is actually registered.
func Locale(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "locale is required"}
	}
	if !bcp47Re.MatchString(value) {
		return &ValidationError{Field: field, Message: "invalid BCP 47 locale format"}
	}
	return nil
}

// iso3166Re matches ISO 3166-1 alpha-2 country codes.
var iso3166Re = regexp.MustCompile(`^[A-Z]{2}$`)

// CountryCode validates that value is a two-letter uppercase ISO 3166-1
// alpha-2 country code (e.g. "US", "DE", "JP"). It checks format only.
func CountryCode(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "country code is required"}
	}
	if !iso3166Re.MatchString(strings.ToUpper(value)) {
		return &ValidationError{Field: field, Message: "must be a two-letter ISO 3166-1 alpha-2 country code"}
	}
	return nil
}

// iso4217Re matches ISO 4217 currency codes.
var iso4217Re = regexp.MustCompile(`^[A-Z]{3}$`)

// CurrencyCode validates that value is a three-letter uppercase ISO 4217
// currency code (e.g. "USD", "EUR", "JPY"). It checks format only.
func CurrencyCode(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "currency code is required"}
	}
	if !iso4217Re.MatchString(strings.ToUpper(value)) {
		return &ValidationError{Field: field, Message: "must be a three-letter ISO 4217 currency code"}
	}
	return nil
}

// iso639Re matches ISO 639-1 (2-letter) and ISO 639-2 (3-letter) codes.
var iso639Re = regexp.MustCompile(`^[a-zA-Z]{2,3}$`)

// LanguageCode validates that value is a two- or three-letter ISO 639
// language code (e.g. "en", "fra"). It checks format only.
func LanguageCode(field, value string) error {
	if value == "" {
		return &ValidationError{Field: field, Message: "language code is required"}
	}
	if !iso639Re.MatchString(value) {
		return &ValidationError{Field: field, Message: "must be a 2 or 3 letter ISO 639 language code"}
	}
	return nil
}
