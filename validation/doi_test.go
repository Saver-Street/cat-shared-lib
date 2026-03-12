package validation

import "testing"

func TestDOIValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
	}{
		{"simple", "10.1000/xyz123"},
		{"real_article", "10.1038/nphys1170"},
		{"multi_segment", "10.1000.10/test"},
		{"complex_suffix", "10.1002/(SICI)1097-4636"},
		{"with_dots", "10.12345.6789/test.data"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := DOI("doi", tc.value); err != nil {
				t.Errorf("DOI(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestDOIWithURLPrefix(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
	}{
		{"https_doi_org", "https://doi.org/10.1000/xyz123"},
		{"http_doi_org", "http://doi.org/10.1000/xyz123"},
		{"https_dx", "https://dx.doi.org/10.1000/xyz123"},
		{"http_dx", "http://dx.doi.org/10.1000/xyz123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := DOI("doi", tc.value); err != nil {
				t.Errorf("DOI(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestDOIWithSpaces(t *testing.T) {
	t.Parallel()
	if err := DOI("doi", "  10.1000/xyz123  "); err != nil {
		t.Errorf("DOI with spaces should be valid: %v", err)
	}
}

func TestDOIInvalid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"no_prefix", "11.1000/xyz"},
		{"short_registrant", "10.12/xyz"},
		{"no_slash", "10.1000xyz"},
		{"no_suffix", "10.1000/"},
		{"whitespace_suffix", "10.1000/ "},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := DOI("doi", tc.value); err == nil {
				t.Errorf("DOI(%q) should fail", tc.value)
			}
		})
	}
}

func TestDOIErrorField(t *testing.T) {
	t.Parallel()
	err := DOI("ref", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Field != "ref" {
		t.Errorf("Field = %q, want %q", err.Field, "ref")
	}
}

func BenchmarkDOI(b *testing.B) {
	for b.Loop() {
		DOI("doi", "10.1038/nphys1170")
	}
}

func FuzzDOI(f *testing.F) {
	f.Add("10.1000/xyz123")
	f.Add("https://doi.org/10.1000/xyz")
	f.Add("")
	f.Add("bad")
	f.Fuzz(func(t *testing.T, s string) {
		DOI("doi", s)
	})
}
