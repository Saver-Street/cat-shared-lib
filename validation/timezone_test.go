package validation

import "testing"

func TestTimezone(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"UTC", "UTC", true},
		{"America/New_York", "America/New_York", true},
		{"Europe/London", "Europe/London", true},
		{"Asia/Tokyo", "Asia/Tokyo", true},
		{"Local", "Local", true},
		{"invalid", "Invalid/Zone", false},
		{"empty", "", false},
		{"random", "not_a_timezone", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Timezone("tz", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("Timezone(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("Timezone(%q) = nil, want error", tt.value)
			}
		})
	}
}

func TestTimezoneOffset(t *testing.T) {
	tests := []struct {
		name  string
		value string
		ok    bool
	}{
		{"Z", "Z", true},
		{"positive", "+05:30", true},
		{"negative", "-05:00", true},
		{"zero", "+00:00", true},
		{"max positive", "+14:00", true},
		{"empty", "", false},
		{"no sign", "005:30", false},
		{"no colon", "+05300", false},
		{"too long", "+05:300", false},
		{"too short", "+05:3", false},
		{"hour out of range", "+15:00", false},
		{"minute out of range", "+05:60", false},
		{"14 with minutes", "+14:01", false},
		{"non-digit", "+ab:cd", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TimezoneOffset("offset", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("TimezoneOffset(%q) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("TimezoneOffset(%q) = nil, want error", tt.value)
			}
		})
	}
}

func BenchmarkTimezone(b *testing.B) {
	for b.Loop() {
		Timezone("tz", "America/New_York")
	}
}

func BenchmarkTimezoneOffset(b *testing.B) {
	for b.Loop() {
		TimezoneOffset("offset", "+05:30")
	}
}

func FuzzTimezone(f *testing.F) {
	f.Add("UTC")
	f.Add("America/New_York")
	f.Add("")
	f.Add("Invalid")

	f.Fuzz(func(t *testing.T, s string) {
		_ = Timezone("tz", s)
	})
}

func FuzzTimezoneOffset(f *testing.F) {
	f.Add("+05:30")
	f.Add("-05:00")
	f.Add("Z")
	f.Add("")
	f.Add("bad")

	f.Fuzz(func(t *testing.T, s string) {
		_ = TimezoneOffset("offset", s)
	})
}
