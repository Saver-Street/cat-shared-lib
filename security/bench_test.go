package security

import "testing"

func BenchmarkContainsSuspiciousInput_Script(b *testing.B) {
	for b.Loop() {
		ContainsSuspiciousInput("<script>alert('xss')</script>")
	}
}

func BenchmarkRedactPII_Nested(b *testing.B) {
	data := map[string]any{
		"user": map[string]any{
			"email":    "user@example.com",
			"password": "secret",
			"profile": map[string]any{
				"ssn":   "123-45-6789",
				"phone": "+1-555-0100",
			},
		},
		"metadata": map[string]any{
			"ip":         "192.168.1.1",
			"user_agent": "Mozilla/5.0",
		},
	}
	for b.Loop() {
		RedactPII(data)
	}
}

func BenchmarkTruncateForLog_Short(b *testing.B) {
	for b.Loop() {
		TruncateForLog("short", 50)
	}
}
