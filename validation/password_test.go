package validation

import (
	"strings"
	"testing"
)

func TestPasswordStrength_Valid(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"standard", "P@ssw0rd!"},
		{"long", "MyStr0ng#Password123"},
		{"minimal", "Aa1!aaaa"},
		{"unicode", "Пароль1!"},
		{"symbols", "Ab1~`!@#$%"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := PasswordStrength("password", tt.password, 8); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestPasswordStrength_TooShort(t *testing.T) {
	err := PasswordStrength("password", "Ab1!", 8)
	if err == nil {
		t.Fatal("expected error for short password")
	}
	if !strings.Contains(err.Error(), "at least 8 characters") {
		t.Errorf("error = %q, want mention of 8 characters", err)
	}
}

func TestPasswordStrength_NoUppercase(t *testing.T) {
	err := PasswordStrength("password", "abcdef1!", 8)
	if err == nil {
		t.Fatal("expected error for no uppercase")
	}
	if !strings.Contains(err.Error(), "uppercase") {
		t.Errorf("error = %q, want mention of uppercase", err)
	}
}

func TestPasswordStrength_NoLowercase(t *testing.T) {
	err := PasswordStrength("password", "ABCDEF1!", 8)
	if err == nil {
		t.Fatal("expected error for no lowercase")
	}
	if !strings.Contains(err.Error(), "lowercase") {
		t.Errorf("error = %q, want mention of lowercase", err)
	}
}

func TestPasswordStrength_NoDigit(t *testing.T) {
	err := PasswordStrength("password", "Abcdefgh!", 8)
	if err == nil {
		t.Fatal("expected error for no digit")
	}
	if !strings.Contains(err.Error(), "digit") {
		t.Errorf("error = %q, want mention of digit", err)
	}
}

func TestPasswordStrength_NoSpecial(t *testing.T) {
	err := PasswordStrength("password", "Abcdefg1", 8)
	if err == nil {
		t.Fatal("expected error for no special character")
	}
	if !strings.Contains(err.Error(), "special") {
		t.Errorf("error = %q, want mention of special", err)
	}
}

func TestPasswordStrength_Empty(t *testing.T) {
	err := PasswordStrength("password", "", 8)
	if err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestPasswordStrength_MinLenZero(t *testing.T) {
	// With minLen=0, only complexity rules apply.
	err := PasswordStrength("password", "Aa1!", 0)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordStrength_ExactMinLength(t *testing.T) {
	err := PasswordStrength("password", "Aa1!aaaa", 8) // exactly 8 chars
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordStrength_FieldName(t *testing.T) {
	err := PasswordStrength("new_password", "short", 8)
	if err == nil {
		t.Fatal("expected error")
	}
	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("expected *ValidationError, got %T", err)
	}
	if ve.Field != "new_password" {
		t.Errorf("Field = %q, want %q", ve.Field, "new_password")
	}
}

func TestPasswordMatch_Match(t *testing.T) {
	if err := PasswordMatch("confirm", "P@ssw0rd", "P@ssw0rd"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordMatch_Mismatch(t *testing.T) {
	err := PasswordMatch("confirm", "P@ssw0rd", "different")
	if err == nil {
		t.Fatal("expected error for mismatched passwords")
	}
	if !strings.Contains(err.Error(), "do not match") {
		t.Errorf("error = %q, want mention of match", err)
	}
}

func TestPasswordMatch_Empty(t *testing.T) {
	if err := PasswordMatch("confirm", "", ""); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPasswordStrength_WithValidator(t *testing.T) {
	v := NewValidator()
	v.Check(PasswordStrength("password", "P@ssw0rd!", 8))
	v.Check(PasswordMatch("confirm", "P@ssw0rd!", "P@ssw0rd!"))
	if !v.Valid() {
		t.Errorf("unexpected errors: %v", v.Errors())
	}
}

func TestPasswordStrength_WithValidatorFailure(t *testing.T) {
	v := NewValidator()
	v.Check(PasswordStrength("password", "weak", 8))
	v.Check(PasswordMatch("confirm", "weak", "different"))
	if v.Valid() {
		t.Error("expected validation failure")
	}
	if len(v.Errors()) != 2 {
		t.Errorf("errors count = %d, want 2", len(v.Errors()))
	}
}

func BenchmarkPasswordStrength_Valid(b *testing.B) {
	for b.Loop() {
		_ = PasswordStrength("password", "P@ssw0rd!Str0ng", 8)
	}
}

func BenchmarkPasswordStrength_Invalid(b *testing.B) {
	for b.Loop() {
		_ = PasswordStrength("password", "weak", 8)
	}
}

func BenchmarkPasswordMatch(b *testing.B) {
	for b.Loop() {
		_ = PasswordMatch("confirm", "P@ssw0rd!", "P@ssw0rd!")
	}
}

func FuzzPasswordStrength(f *testing.F) {
	f.Add("P@ssw0rd!", 8)
	f.Add("", 0)
	f.Add("abcdefgh", 8)
	f.Add("ABCDEFGH", 8)
	f.Add("12345678", 8)
	f.Add("!@#$%^&*", 8)
	f.Add("Aa1!", 4)
	f.Add("Пароль1!", 8)
	f.Add("\x00\xff\n\t", 1)
	f.Add("A" + strings.Repeat("a", 999) + "1!", 1000)

	f.Fuzz(func(t *testing.T, password string, minLen int) {
		if minLen < 0 {
			minLen = 0
		}
		err := PasswordStrength("password", password, minLen)
		// Must not panic. If err is nil, verify the password actually meets requirements.
		if err == nil && len(password) < minLen {
			t.Errorf("accepted password shorter than minLen: len=%d, minLen=%d", len(password), minLen)
		}
	})
}
