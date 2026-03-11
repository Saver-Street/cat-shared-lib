package validation

import (
	"errors"
	"testing"
)

func TestCronExpressionValid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"every minute", "* * * * *"},
		{"specific time", "30 2 * * *"},
		{"range", "0-30 * * * *"},
		{"step", "*/5 * * * *"},
		{"list", "1,15,30 * * * *"},
		{"complex", "0,30 9-17 * 1-6 1-5"},
		{"range with step", "0-59/2 * * * *"},
		{"day of week 0", "0 0 * * 0"},
		{"day of week 6", "0 0 * * 6"},
		{"all specific", "5 4 3 2 1"},
		{"max values", "59 23 31 12 6"},
		{"min values", "0 0 1 1 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := CronExpression("schedule", tt.value); err != nil {
				t.Errorf("CronExpression(%q) = %v; want nil", tt.value, err)
			}
		})
	}
}

func TestCronExpressionInvalid(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
	}{
		{"empty", ""},
		{"too few fields", "* * *"},
		{"too many fields", "* * * * * *"},
		{"invalid minute", "60 * * * *"},
		{"invalid hour", "* 24 * * *"},
		{"invalid day", "* * 0 * *"},
		{"invalid day high", "* * 32 * *"},
		{"invalid month", "* * * 0 *"},
		{"invalid month high", "* * * 13 *"},
		{"invalid dow", "* * * * 7"},
		{"bad step", "*/0 * * * *"},
		{"bad step text", "*/abc * * * *"},
		{"bad range start", "abc-5 * * * *"},
		{"bad range end", "0-abc * * * *"},
		{"range inverted", "30-5 * * * *"},
		{"bad value text", "abc * * * *"},
		{"negative", "-1 * * * *"},
		{"range start too high", "60-60 * * * *"},
		{"range end too high", "0-60 * * * *"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CronExpression("schedule", tt.value)
			if err == nil {
				t.Errorf("CronExpression(%q) = nil; want error", tt.value)
			}
		})
	}
}

func TestCronExpressionFieldName(t *testing.T) {
	t.Parallel()
	err := CronExpression("cron", "bad")
	if err == nil {
		t.Fatal("expected error")
	}
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("error type = %T; want *ValidationError", err)
	}
	if ve.Field != "cron" {
		t.Errorf("Field = %q; want cron", ve.Field)
	}
}

func TestCronExpressionListInvalid(t *testing.T) {
	t.Parallel()
	err := CronExpression("f", "1,60 * * * *")
	if err == nil {
		t.Error("expected error for list with out-of-range value")
	}
}

func BenchmarkCronExpression(b *testing.B) {
	for range b.N {
		_ = CronExpression("s", "*/5 0-23 1,15 1-12 0-6")
	}
}

func FuzzCronExpression(f *testing.F) {
	f.Add("* * * * *")
	f.Add("0 0 1 1 0")
	f.Add("bad")
	f.Add("")
	f.Fuzz(func(t *testing.T, value string) {
		_ = CronExpression("f", value)
	})
}
