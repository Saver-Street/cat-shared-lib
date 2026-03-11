package validation

import (
	"testing"
)

func TestSemver_Valid(t *testing.T) {
	valid := []string{
		"0.0.0",
		"1.0.0",
		"1.2.3",
		"v1.2.3",
		"0.1.0",
		"10.20.30",
		"1.0.0-alpha",
		"1.0.0-alpha.1",
		"1.0.0-0.3.7",
		"1.0.0-x.7.z.92",
		"1.0.0+build.1",
		"1.0.0-beta+exp.sha.5114f85",
		"v2.0.0-rc.1",
		"1.0.0-alpha.beta",
	}
	for _, v := range valid {
		t.Run(v, func(t *testing.T) {
			if err := Semver("version", v); err != nil {
				t.Errorf("Semver(%q) = %v, want nil", v, err)
			}
		})
	}
}

func TestSemver_Invalid(t *testing.T) {
	invalid := []string{
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"a.b.c",
		"01.0.0",
		"1.02.0",
		"1.0.03",
		"1.0.0-",
		"1.0.0+",
		"v",
		"latest",
		"1.0.0-alpha..1",
	}
	for _, v := range invalid {
		t.Run(v, func(t *testing.T) {
			if err := Semver("version", v); err == nil {
				t.Errorf("Semver(%q) = nil, want error", v)
			}
		})
	}
}

func TestSemver_EmptyRequired(t *testing.T) {
	err := Semver("version", "")
	if err == nil {
		t.Fatal("expected error for empty value")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if ve.Field != "version" {
		t.Errorf("field = %q, want %q", ve.Field, "version")
	}
}

func TestSemver_WhitespaceTrimmed(t *testing.T) {
	if err := Semver("v", "  1.2.3  "); err != nil {
		t.Errorf("expected trimmed version to be valid: %v", err)
	}
}

func TestParseSemver(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		pre     string
		build   string
		wantErr bool
	}{
		{"1.2.3", 1, 2, 3, "", "", false},
		{"v0.1.0", 0, 1, 0, "", "", false},
		{"1.0.0-alpha.1", 1, 0, 0, "alpha.1", "", false},
		{"1.0.0+build.42", 1, 0, 0, "", "build.42", false},
		{"2.0.0-rc.1+sha.abc", 2, 0, 0, "rc.1", "sha.abc", false},
		{"invalid", 0, 0, 0, "", "", true},
		{"", 0, 0, 0, "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			p, err := ParseSemver(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseSemver(%q) = nil error, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseSemver(%q) = %v", tt.input, err)
			}
			if p.Major != tt.major || p.Minor != tt.minor || p.Patch != tt.patch {
				t.Errorf("version = %d.%d.%d, want %d.%d.%d", p.Major, p.Minor, p.Patch, tt.major, tt.minor, tt.patch)
			}
			if p.Prerelease != tt.pre {
				t.Errorf("prerelease = %q, want %q", p.Prerelease, tt.pre)
			}
			if p.Build != tt.build {
				t.Errorf("build = %q, want %q", p.Build, tt.build)
			}
		})
	}
}

func TestCompareSemver(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "2.0.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0-alpha", "1.0.0-alpha", 0},
		{"1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-1", "1.0.0-2", -1},
		{"1.0.0-1", "1.0.0-alpha", -1},
		{"1.0.0-alpha", "1.0.0-1", 1},
		{"v1.0.0", "1.0.0", 0},
		// Build metadata is ignored.
		{"1.0.0+build1", "1.0.0+build2", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := CompareSemver(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("CompareSemver(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestSemverMinVersion(t *testing.T) {
	tests := []struct {
		value   string
		min     string
		wantErr bool
	}{
		{"2.0.0", "1.0.0", false},
		{"1.0.0", "1.0.0", false},
		{"1.0.0", "2.0.0", true},
		{"1.0.0-alpha", "1.0.0", true},
		{"1.0.0", "1.0.0-alpha", false},
		{"v1.2.3", "v1.2.0", false},
		{"", "1.0.0", true},
		{"1.0.0", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.value+"_min_"+tt.min, func(t *testing.T) {
			err := SemverMinVersion("version", tt.value, tt.min)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q >= %q", tt.value, tt.min)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func BenchmarkSemver(b *testing.B) {
	for b.Loop() {
		Semver("version", "1.2.3-beta.1+build.42")
	}
}

func BenchmarkParseSemver(b *testing.B) {
	for b.Loop() {
		ParseSemver("v2.0.0-rc.1+sha.abc")
	}
}

func BenchmarkCompareSemver(b *testing.B) {
	for b.Loop() {
		CompareSemver("1.0.0-alpha.1", "1.0.0-beta.2")
	}
}

func FuzzSemver(f *testing.F) {
	f.Add("1.0.0")
	f.Add("v2.0.0-beta.1+build.42")
	f.Add("")
	f.Add("abc")
	f.Add("1.2")
	f.Add("01.0.0")
	f.Add("1.0.0-alpha..1")

	f.Fuzz(func(t *testing.T, input string) {
		// Must not panic.
		err := Semver("field", input)
		if err == nil {
			// Valid semver should also parse successfully.
			p, parseErr := ParseSemver(input)
			if parseErr != nil {
				t.Errorf("Semver valid but ParseSemver failed: %v", parseErr)
			}
			// Self-comparison must be 0.
			if cmp := CompareSemver(input, input); cmp != 0 {
				t.Errorf("CompareSemver(%q, %q) = %d, want 0", input, input, cmp)
			}
			_ = p
		}
	})
}

func FuzzCompareSemver(f *testing.F) {
	f.Add("1.0.0", "2.0.0")
	f.Add("1.0.0-alpha", "1.0.0")
	f.Add("invalid", "1.0.0")

	f.Fuzz(func(t *testing.T, a, b string) {
		ab := CompareSemver(a, b)
		ba := CompareSemver(b, a)
		// Antisymmetry: sign(compare(a,b)) == -sign(compare(b,a))
		if ab != -ba && !(ab == 0 && ba == 0) {
			t.Errorf("antisymmetry violated: CompareSemver(%q,%q)=%d, reverse=%d", a, b, ab, ba)
		}
	})
}
