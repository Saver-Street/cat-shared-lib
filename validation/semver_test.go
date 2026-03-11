package validation

import "testing"

func TestSemVer(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"basic", "1.2.3", true},
		{"with v", "v1.2.3", true},
		{"with prerelease", "1.0.0-alpha", true},
		{"with build", "1.0.0+build.1", true},
		{"with both", "1.0.0-beta.1+build.123", true},
		{"zero version", "0.0.0", true},
		{"large numbers", "100.200.300", true},
		{"missing patch", "1.2", false},
		{"extra part", "1.2.3.4", false},
		{"empty", "", false},
		{"letters", "abc", false},
		{"negative", "-1.0.0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SemVer("version", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("SemVer(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("SemVer(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestParseSemVer(t *testing.T) {
	tests := []struct {
		name  string
		input string
		major int
		minor int
		patch int
		pre   string
		build string
		ok    bool
	}{
		{"basic", "1.2.3", 1, 2, 3, "", "", true},
		{"with v", "v1.0.0", 1, 0, 0, "", "", true},
		{"prerelease", "2.0.0-rc.1", 2, 0, 0, "rc.1", "", true},
		{"build", "1.0.0+20230101", 1, 0, 0, "", "20230101", true},
		{"both", "3.1.4-beta+build", 3, 1, 4, "beta", "build", true},
		{"invalid", "bad", 0, 0, 0, "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSemVer(tt.input)
			if tt.ok {
				if err != nil {
					t.Fatalf("ParseSemVer(%q) error: %v", tt.input, err)
				}
				if got.Major != tt.major || got.Minor != tt.minor || got.Patch != tt.patch {
					t.Fatalf("version = %d.%d.%d, want %d.%d.%d", got.Major, got.Minor, got.Patch, tt.major, tt.minor, tt.patch)
				}
				if got.Prerelease != tt.pre {
					t.Fatalf("prerelease = %q, want %q", got.Prerelease, tt.pre)
				}
				if got.Build != tt.build {
					t.Fatalf("build = %q, want %q", got.Build, tt.build)
				}
			} else {
				if err == nil {
					t.Fatalf("ParseSemVer(%q) = nil error, want error", tt.input)
				}
			}
		})
	}
}

func TestSemVerMinVersion(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		minVer string
		ok     bool
	}{
		{"above min", "2.0.0", "1.0.0", true},
		{"equal", "1.0.0", "1.0.0", true},
		{"below min", "0.9.0", "1.0.0", false},
		{"minor above", "1.5.0", "1.2.0", true},
		{"minor below", "1.1.0", "1.2.0", false},
		{"patch above", "1.0.5", "1.0.3", true},
		{"patch below", "1.0.1", "1.0.3", false},
		{"invalid value", "bad", "1.0.0", false},
		{"invalid min", "1.0.0", "bad", false},
		{"prerelease below release", "1.0.0-alpha", "1.0.0", false},
		{"release above prerelease", "1.0.0", "1.0.0-alpha", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SemVerMinVersion("version", tt.value, tt.minVer)
			if tt.ok && err != nil {
				t.Fatalf("SemVerMinVersion(%q, %q) = %v, want nil", tt.value, tt.minVer, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("SemVerMinVersion(%q, %q) = nil, want error", tt.value, tt.minVer)
			}
		})
	}
}

func TestComparePre(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "alpha", 1},
		{"alpha", "", -1},
		{"alpha", "beta", -1},
		{"beta", "alpha", 1},
		{"alpha", "alpha", 0},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			got := comparePre(tt.a, tt.b)
			if (tt.want < 0 && got >= 0) || (tt.want > 0 && got <= 0) || (tt.want == 0 && got != 0) {
				t.Fatalf("comparePre(%q, %q) = %d, want sign of %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func BenchmarkSemVer(b *testing.B) {
	for b.Loop() {
		SemVer("version", "1.2.3-beta.1+build.456")
	}
}

func BenchmarkParseSemVer(b *testing.B) {
	for b.Loop() {
		ParseSemVer("v1.2.3-beta.1+build.456")
	}
}

func FuzzSemVer(f *testing.F) {
	f.Add("1.2.3")
	f.Add("v0.0.0")
	f.Add("")
	f.Add("bad")
	f.Add("1.0.0-alpha+build")

	f.Fuzz(func(t *testing.T, s string) {
		_ = SemVer("v", s)
	})
}

func FuzzParseSemVer(f *testing.F) {
	f.Add("1.2.3")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_, _ = ParseSemVer(s)
	})
}
