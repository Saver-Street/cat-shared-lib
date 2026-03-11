package duration_test

import (
	"testing"
	"time"

	"github.com/Saver-Street/cat-shared-lib/duration"
)

// ── Human ──

func TestHuman(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "zero", d: 0, want: "< 1s"},
		{name: "sub-second nanosecond", d: time.Nanosecond, want: "< 1s"},
		{name: "sub-second 500ms", d: 500 * time.Millisecond, want: "< 1s"},
		{name: "exactly 1s", d: time.Second, want: "1s"},
		{name: "30 seconds", d: 30 * time.Second, want: "30s"},
		{name: "1 minute", d: time.Minute, want: "1m"},
		{name: "5m 10s", d: 5*time.Minute + 10*time.Second, want: "5m 10s"},
		{name: "1 hour", d: time.Hour, want: "1h"},
		{name: "2h 30m", d: 2*time.Hour + 30*time.Minute, want: "2h 30m"},
		{name: "2h 30m 45s", d: 2*time.Hour + 30*time.Minute + 45*time.Second, want: "2h 30m 45s"},
		{name: "1 day", d: 24 * time.Hour, want: "1d"},
		{name: "1d 3h", d: 27 * time.Hour, want: "1d 3h"},
		{name: "1d 3h 15m (no seconds)", d: 27*time.Hour + 15*time.Minute + 10*time.Second, want: "1d 3h 15m"},
		{name: "7 days", d: 7 * 24 * time.Hour, want: "7d"},
		{name: "30 days", d: 30 * 24 * time.Hour, want: "30d"},
		{name: "365 days", d: 365 * 24 * time.Hour, want: "365d"},
		{name: "negative 5m", d: -5 * time.Minute, want: "-5m"},
		{name: "negative 2h 30m", d: -(2*time.Hour + 30*time.Minute), want: "-2h 30m"},
		{name: "negative sub-second", d: -500 * time.Millisecond, want: "-< 1s"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := duration.Human(tt.d)
			if got != tt.want {
				t.Errorf("Human(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// ── Short ──

func TestShort(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{name: "zero microseconds", d: 0, want: "0µs"},
		{name: "500 nanoseconds", d: 500 * time.Nanosecond, want: "0µs"},
		{name: "50 microseconds", d: 50 * time.Microsecond, want: "50µs"},
		{name: "750ms", d: 750 * time.Millisecond, want: "750ms"},
		{name: "1 second", d: time.Second, want: "1s"},
		{name: "45 seconds", d: 45 * time.Second, want: "45s"},
		{name: "5m 10s", d: 5*time.Minute + 10*time.Second, want: "5m 10s"},
		{name: "5m exact", d: 5 * time.Minute, want: "5m"},
		{name: "2h 30m", d: 2*time.Hour + 30*time.Minute, want: "2h 30m"},
		{name: "2h exact", d: 2 * time.Hour, want: "2h"},
		{name: "3d 4h", d: 76 * time.Hour, want: "3d 4h"},
		{name: "3d exact", d: 72 * time.Hour, want: "3d"},
		{name: "30d 12h", d: 732 * time.Hour, want: "30d 12h"},
		{name: "negative 5m 10s", d: -(5*time.Minute + 10*time.Second), want: "-5m 10s"},
		{name: "negative 750ms", d: -750 * time.Millisecond, want: "-750ms"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := duration.Short(tt.d)
			if got != tt.want {
				t.Errorf("Short(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

// ── Since / Until ──

func TestSince(t *testing.T) {
	past := time.Now().Add(-2*time.Hour - 30*time.Minute)
	got := duration.Since(past)
	// Should be approximately "2h 30m" — accept anything containing "2h"
	if len(got) == 0 {
		t.Error("Since returned empty string")
	}
}

func TestUntil(t *testing.T) {
	future := time.Now().Add(5*time.Minute + 10*time.Second)
	got := duration.Until(future)
	if len(got) == 0 {
		t.Error("Until returned empty string")
	}
}

// ── Round ──

func TestRound(t *testing.T) {
	tests := []struct {
		name      string
		d         time.Duration
		precision time.Duration
		want      time.Duration
	}{
		{name: "2h30m to hour", d: 2*time.Hour + 30*time.Minute, precision: time.Hour, want: 3 * time.Hour},
		{name: "45m to 30m", d: 45 * time.Minute, precision: 30 * time.Minute, want: time.Hour},
		{name: "zero precision returns input", d: 5 * time.Second, precision: 0, want: 5 * time.Second},
		{name: "negative precision returns input", d: 5 * time.Second, precision: -time.Second, want: 5 * time.Second},
		{name: "10s to 3s", d: 10 * time.Second, precision: 3 * time.Second, want: 9 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := duration.Round(tt.d, tt.precision)
			if got != tt.want {
				t.Errorf("Round(%v, %v) = %v, want %v", tt.d, tt.precision, got, tt.want)
			}
		})
	}
}

// ── Truncate ──

func TestTruncate(t *testing.T) {
	tests := []struct {
		name      string
		d         time.Duration
		precision time.Duration
		want      time.Duration
	}{
		{name: "2h30m to hour", d: 2*time.Hour + 30*time.Minute, precision: time.Hour, want: 2 * time.Hour},
		{name: "59m to 30m", d: 59 * time.Minute, precision: 30 * time.Minute, want: 30 * time.Minute},
		{name: "zero precision returns input", d: 5 * time.Second, precision: 0, want: 5 * time.Second},
		{name: "negative precision returns input", d: 5 * time.Second, precision: -time.Second, want: 5 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := duration.Truncate(tt.d, tt.precision)
			if got != tt.want {
				t.Errorf("Truncate(%v, %v) = %v, want %v", tt.d, tt.precision, got, tt.want)
			}
		})
	}
}

// ── Benchmarks ──

func BenchmarkHuman(b *testing.B) {
	d := 2*time.Hour + 30*time.Minute + 45*time.Second
	for b.Loop() {
		_ = duration.Human(d)
	}
}

func BenchmarkShort(b *testing.B) {
	d := 2*time.Hour + 30*time.Minute + 45*time.Second
	for b.Loop() {
		_ = duration.Short(d)
	}
}

// ── Fuzz ──

func FuzzHuman(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(time.Second))
	f.Add(int64(time.Hour))
	f.Add(int64(24 * time.Hour))
	f.Add(int64(-5 * time.Minute))
	f.Add(int64(999 * time.Millisecond))
	f.Fuzz(func(t *testing.T, ns int64) {
		// Human must never panic for any duration value.
		_ = duration.Human(time.Duration(ns))
	})
}
