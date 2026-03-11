package duration

import (
	"fmt"
	"time"
)

// Human formats a time.Duration as a human-readable string.
// Examples: "2h 30m", "5m 10s", "1d 3h", "< 1s"
func Human(d time.Duration) string {
	if d < 0 {
		return "-" + Human(-d)
	}
	if d < time.Second {
		return "< 1s"
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}
	if seconds > 0 && days == 0 { // skip seconds when showing days
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	if len(parts) == 0 {
		return "< 1s"
	}
	// Join parts with space
	result := parts[0]
	for _, p := range parts[1:] {
		result += " " + p
	}
	return result
}

// Short formats a duration with at most 2 units for brevity.
// Examples: "2h 30m", "5m 10s", "3d 4h", "45s"
func Short(d time.Duration) string {
	if d < 0 {
		return "-" + Short(-d)
	}
	if d < time.Second {
		if d < time.Millisecond {
			return fmt.Sprintf("%dµs", d.Microseconds())
		}
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	switch {
	case days > 0:
		if hours > 0 {
			return fmt.Sprintf("%dd %dh", days, hours)
		}
		return fmt.Sprintf("%dd", days)
	case hours > 0:
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	case minutes > 0:
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// Since returns a human-readable string of the duration since t.
// Commonly used for "time ago" displays.
func Since(t time.Time) string {
	return Human(time.Since(t))
}

// Until returns a human-readable string of the duration until t.
func Until(t time.Time) string {
	return Human(time.Until(t))
}

// Round rounds d to the nearest unit of precision.
// Examples: Round(2h30m, 1h) → 3h, Round(45m, 30m) → 30m
func Round(d, precision time.Duration) time.Duration {
	if precision <= 0 {
		return d
	}
	return d.Round(precision)
}

// Truncate truncates d to the given precision.
// Examples: Truncate(2h30m, 1h) → 2h
func Truncate(d, precision time.Duration) time.Duration {
	if precision <= 0 {
		return d
	}
	return d.Truncate(precision)
}
