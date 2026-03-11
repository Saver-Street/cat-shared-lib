package validation

import (
	"regexp"
	"strconv"
	"strings"
)

// semverRe matches semantic version strings (v prefix optional).
var semverRe = regexp.MustCompile(`^v?(\d+)\.(\d+)\.(\d+)(?:-([0-9A-Za-z\-.]+))?(?:\+([0-9A-Za-z\-.]+))?$`)

// SemVer validates that value is a valid semantic version string (e.g.
// "1.2.3", "v1.0.0-beta+build.1").
func SemVer(field, value string) error {
	if !semverRe.MatchString(value) {
		return &ValidationError{Field: field, Message: "invalid semantic version"}
	}
	return nil
}

// SemVerParts represents the parsed components of a semantic version.
type SemVerParts struct {
	Major      int
	Minor      int
	Patch      int
	Prerelease string
	Build      string
}

// ParseSemVer parses a semantic version string into its components.
// Returns an error if the input is not a valid semver.
func ParseSemVer(value string) (SemVerParts, error) {
	matches := semverRe.FindStringSubmatch(value)
	if matches == nil {
		return SemVerParts{}, &ValidationError{Field: "version", Message: "invalid semantic version"}
	}
	major, _ := strconv.Atoi(matches[1])
	minor, _ := strconv.Atoi(matches[2])
	patch, _ := strconv.Atoi(matches[3])

	return SemVerParts{
		Major:      major,
		Minor:      minor,
		Patch:      patch,
		Prerelease: matches[4],
		Build:      matches[5],
	}, nil
}

// SemVerMinVersion validates that value is a semantic version >= the
// specified minimum version.
func SemVerMinVersion(field, value, minVersion string) error {
	if err := SemVer(field, value); err != nil {
		return err
	}
	cur, _ := ParseSemVer(value)
	min, err := ParseSemVer(minVersion)
	if err != nil {
		return &ValidationError{Field: field, Message: "invalid minimum version"}
	}
	if compareSemVer(cur, min) < 0 {
		return &ValidationError{Field: field, Message: "version below minimum " + minVersion}
	}
	return nil
}

// compareSemVer compares two versions.  Returns negative if a < b, 0 if
// equal, positive if a > b.  Pre-release versions are compared
// lexicographically; a version without pre-release is greater than one
// with pre-release at the same major.minor.patch.
func compareSemVer(a, b SemVerParts) int {
	if a.Major != b.Major {
		return a.Major - b.Major
	}
	if a.Minor != b.Minor {
		return a.Minor - b.Minor
	}
	if a.Patch != b.Patch {
		return a.Patch - b.Patch
	}
	// Pre-release comparison per semver spec.
	return comparePre(a.Prerelease, b.Prerelease)
}

func comparePre(a, b string) int {
	if a == b {
		return 0
	}
	// No pre-release is higher than having one.
	if a == "" {
		return 1
	}
	if b == "" {
		return -1
	}
	return strings.Compare(a, b)
}
