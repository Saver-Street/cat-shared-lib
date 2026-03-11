package validation

import "testing"

func TestISSNValid(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		value string
	}{
		{"nature", "0028-0836"},
		{"science", "0036-8075"},
		{"check_x", "0009-000X"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := ISSN("issn", tc.value); err != nil {
				t.Errorf("ISSN(%q) = %v, want nil", tc.value, err)
			}
		})
	}
}

func TestISSNValidLowercaseX(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "0009-000x"); err != nil {
		t.Errorf("lowercase x check digit should be valid: %v", err)
	}
}

func TestISSNValidWithSpaces(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "  0028-0836  "); err != nil {
		t.Errorf("ISSN with spaces should be valid: %v", err)
	}
}

func TestISSNInvalidLength(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "0028-083"); err == nil {
		t.Error("short ISSN should fail")
	}
}

func TestISSNMissingHyphen(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "002808361"); err == nil {
		t.Error("ISSN without hyphen should fail")
	}
}

func TestISSNInvalidChar(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "003A-8075"); err == nil {
		t.Error("ISSN with letter in digit position should fail")
	}
}

func TestISSNBadCheckDigit(t *testing.T) {
	t.Parallel()
	if err := ISSN("issn", "0028-0837"); err == nil {
		t.Error("ISSN with wrong check digit should fail")
	}
}

func TestISSNErrorField(t *testing.T) {
	t.Parallel()
	err := ISSN("serial", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Field != "serial" {
		t.Errorf("Field = %q, want %q", err.Field, "serial")
	}
}

func BenchmarkISSN(b *testing.B) {
	for b.Loop() {
		ISSN("issn", "0028-0836")
	}
}

func FuzzISSN(f *testing.F) {
	f.Add("0028-0836")
	f.Add("0009-000X")
	f.Add("bad")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		ISSN("issn", s)
	})
}
