package validation

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// semverRe matches a semantic version string per semver.org (v2.0.0).
// Optional leading "v" prefix is stripped before matching.
var semverRe = regexp.MustCompile(
	`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
		`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
		`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
)

// Semver validates that value is a well-formed semantic version string
// (e.g. "1.2.3", "v2.0.0-beta.1+build.42"). An optional leading "v" prefix
// is accepted. Returns a *ValidationError on failure.
func Semver(field, value string) error {
	v := strings.TrimSpace(value)
	if v == "" {
		return &ValidationError{Field: field, Message: field + " is required"}
	}
	v = strings.TrimPrefix(v, "v")
	if !semverRe.MatchString(v) {
		return &ValidationError{Field: field, Message: "invalid semantic version format"}
	}
	return nil
}

// SemverMinVersion validates that value is a semantic version greater than or
// equal to minVersion. Both values may have an optional "v" prefix.
// Pre-release versions are compared lexically per the semver spec.
func SemverMinVersion(field, value, minVersion string) error {
	if err := Semver(field, value); err != nil {
		return err
	}
	if err := Semver(field, minVersion); err != nil {
		return err
	}
	cmp := CompareSemver(value, minVersion)
	if cmp < 0 {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("version must be >= %s", minVersion),
		}
	}
	return nil
}

// ParsedSemver holds the components of a parsed semantic version.
type ParsedSemver struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// ParseSemver parses a semantic version string into its components.
// An optional leading "v" prefix is stripped. Returns a *ValidationError
// if the string is not a valid semver.
func ParseSemver(value string) (ParsedSemver, error) {
	v := strings.TrimSpace(value)
	v = strings.TrimPrefix(v, "v")
	m := semverRe.FindStringSubmatch(v)
	if m == nil {
		return ParsedSemver{}, &ValidationError{Message: "invalid semantic version format"}
	}
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return ParsedSemver{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: m[4],
		Build:      m[5],
	}, nil
}

// CompareSemver compares two semver strings and returns -1, 0, or 1.
// Build metadata is ignored per the semver spec. Invalid versions sort last.
func CompareSemver(a, b string) int {
	pa, errA := ParseSemver(a)
	pb, errB := ParseSemver(b)
	if errA != nil && errB != nil {
		return 0
	}
	if errA != nil {
		return 1
	}
	if errB != nil {
		return -1
	}

	if pa.Major != pb.Major {
		return cmpInt(pa.Major, pb.Major)
	}
	if pa.Minor != pb.Minor {
		return cmpInt(pa.Minor, pb.Minor)
	}
	if pa.Patch != pb.Patch {
		return cmpInt(pa.Patch, pb.Patch)
	}

	// Pre-release has lower precedence than release.
	if pa.Prerelease == "" && pb.Prerelease == "" {
		return 0
	}
	if pa.Prerelease == "" {
		return 1
	}
	if pb.Prerelease == "" {
		return -1
	}
	return comparePrerelease(pa.Prerelease, pb.Prerelease)
}

func comparePrerelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")
	n := len(aParts)
	if len(bParts) < n {
		n = len(bParts)
	}
	for i := 0; i < n; i++ {
		ai, aErr := strconv.Atoi(aParts[i])
		bi, bErr := strconv.Atoi(bParts[i])
		switch {
		case aErr == nil && bErr == nil:
			if ai != bi {
				return cmpInt(ai, bi)
			}
		case aErr == nil:
			return -1 // numeric < alphanumeric
		case bErr == nil:
			return 1
		default:
			if aParts[i] < bParts[i] {
				return -1
			}
			if aParts[i] > bParts[i] {
				return 1
			}
		}
	}
	return cmpInt(len(aParts), len(bParts))
}

func cmpInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
