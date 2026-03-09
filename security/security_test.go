package security

import (
	"testing"
)

func TestContainsSuspiciousInput_SQL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"DROP TABLE users", true},
		{"drop table users", true},
		{"SELECT * FROM users", true},
		{"1; UNION SELECT password FROM users", true},
		{"INSERT INTO users VALUES", true},
		{"DELETE FROM users WHERE 1=1", true},
		{"UPDATE users SET admin = true", true},
		{"normal text", false},
		{"John O'Malley", false},
		{"", false},
		{"   ", false},
	}
	for _, tt := range tests {
		if got := ContainsSuspiciousInput(tt.input); got != tt.want {
			t.Errorf("ContainsSuspiciousInput(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestContainsSuspiciousInput_XSS(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"<script>alert(1)</script>", true},
		{"javascript: alert(1)", true},
		{"onclick=alert(1)", true},
		{"<iframe src=evil.com>", true},
		{"<object data=evil>", true},
		{"<embed src=evil>", true},
		{"<svg onload=alert(1)>", true},
		{"data: text/html,<h1>Hi</h1>", true},
		{"<b>bold text</b>", false},
		{"click here", false},
	}
	for _, tt := range tests {
		if got := ContainsSuspiciousInput(tt.input); got != tt.want {
			t.Errorf("ContainsSuspiciousInput(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRedactPII_Email(t *testing.T) {
	data := map[string]any{"msg": "Contact john@example.com"}
	result := RedactPII(data)
	if result["msg"] != "Contact [EMAIL_REDACTED]" {
		t.Errorf("got %v", result["msg"])
	}
}

func TestRedactPII_Phone(t *testing.T) {
	data := map[string]any{"msg": "Call 555-123-4567"}
	result := RedactPII(data)
	got := result["msg"].(string)
	if got == "Call 555-123-4567" {
		t.Error("phone number not redacted")
	}
}

func TestRedactPII_SSN(t *testing.T) {
	data := map[string]any{"msg": "SSN: 123-45-6789"}
	result := RedactPII(data)
	if result["msg"] != "SSN: [SSN_REDACTED]" {
		t.Errorf("got %v", result["msg"])
	}
}

func TestRedactPII_FieldName(t *testing.T) {
	data := map[string]any{"email": "test@test.com", "name": "John"}
	result := RedactPII(data)
	if result["email"] != "[REDACTED]" {
		t.Errorf("email field not redacted: %v", result["email"])
	}
	if result["name"] != "John" {
		t.Errorf("name field should not be redacted: %v", result["name"])
	}
}

func TestRedactPII_Nested(t *testing.T) {
	data := map[string]any{
		"user": map[string]any{
			"email": "hidden@test.com",
			"name":  "Jane",
		},
	}
	result := RedactPII(data)
	nested := result["user"].(map[string]any)
	if nested["email"] != "[REDACTED]" {
		t.Errorf("nested email not redacted: %v", nested["email"])
	}
}

func TestRedactPII_Array(t *testing.T) {
	data := map[string]any{
		"items": []any{"user@test.com", "plain text"},
	}
	result := RedactPII(data)
	items := result["items"].([]any)
	if items[0] == "user@test.com" {
		t.Error("email in array not redacted")
	}
	if items[1] != "plain text" {
		t.Errorf("plain text changed: %v", items[1])
	}
}

func TestRedactPII_NonString(t *testing.T) {
	data := map[string]any{"count": 42}
	result := RedactPII(data)
	if result["count"] != 42 {
		t.Errorf("int value changed: %v", result["count"])
	}
}

func TestRedactPII_Empty(t *testing.T) {
	result := RedactPII(map[string]any{})
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
}

func TestRedactPII_CaseInsensitiveField(t *testing.T) {
	data := map[string]any{"Password": "secret123"}
	result := RedactPII(data)
	// "Password" has capital P, but piiFields has "password" lowercase
	// The code checks both the original key and lowercase
	if result["Password"] != "[REDACTED]" {
		t.Errorf("Password field not redacted: %v", result["Password"])
	}
}

// --- Benchmarks ---

func BenchmarkContainsSuspiciousInput_Clean(b *testing.B) {
	for b.Loop() {
		ContainsSuspiciousInput("Hello, this is a normal user message with no attacks.")
	}
}

func BenchmarkContainsSuspiciousInput_SQLi(b *testing.B) {
	for b.Loop() {
		ContainsSuspiciousInput("1; DROP TABLE users; --")
	}
}

func BenchmarkRedactPII_Mixed(b *testing.B) {
	data := map[string]any{
		"email":   "user@example.com",
		"name":    "Jane Doe",
		"message": "Call me at 555-123-4567 or email bob@test.com",
		"count":   42,
	}
	for b.Loop() {
		RedactPII(data)
	}
}
