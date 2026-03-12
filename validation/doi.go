package validation

import (
	"regexp"
	"strings"
)

// doiRegexp matches a DOI per the DOI handbook: directory indicator "10."
// followed by a registrant code, a slash, and a suffix of printable characters.
var doiRegexp = regexp.MustCompile(`^10\.\d{4,}(\.\d+)*/[^\s]+$`)

// DOI validates that value is a valid Digital Object Identifier (DOI).
// Accepts bare DOIs (e.g., "10.1000/xyz123") and common URL prefixes
// (https://doi.org/, http://dx.doi.org/).
func DOI(field, value string) *ValidationError {
	v := strings.TrimSpace(value)
	msg := "must be a valid DOI (e.g., 10.1000/xyz123)"

	// Strip common URL prefixes.
	for _, prefix := range []string{
		"https://doi.org/",
		"http://doi.org/",
		"https://dx.doi.org/",
		"http://dx.doi.org/",
	} {
		if strings.HasPrefix(strings.ToLower(v), prefix) {
			v = v[len(prefix):]
			break
		}
	}

	if !doiRegexp.MatchString(v) {
		return &ValidationError{Field: field, Message: msg}
	}
	return nil
}
