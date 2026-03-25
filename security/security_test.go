package security

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
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
		testkit.AssertEqual(t, ContainsSuspiciousInput(tt.input), tt.want)
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
		testkit.AssertEqual(t, ContainsSuspiciousInput(tt.input), tt.want)
	}
}

func TestRedactPII_Email(t *testing.T) {
	data := map[string]any{"msg": "Contact john@example.com"}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["msg"], "Contact [EMAIL_REDACTED]")
}

func TestRedactPII_Phone(t *testing.T) {
	data := map[string]any{"msg": "Call 555-123-4567"}
	result := RedactPII(data)
	got := result["msg"].(string)
	testkit.AssertNotEqual(t, got, "Call 555-123-4567")
}

func TestRedactPII_SSN(t *testing.T) {
	data := map[string]any{"msg": "SSN: 123-45-6789"}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["msg"], "SSN: [SSN_REDACTED]")
}

func TestRedactPII_FieldName(t *testing.T) {
	data := map[string]any{"email": "test@test.com", "name": "John"}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["email"], "[REDACTED]")
	testkit.AssertEqual(t, result["name"], "John")
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
	testkit.AssertEqual(t, nested["email"], "[REDACTED]")
}

func TestRedactPII_Array(t *testing.T) {
	data := map[string]any{
		"items": []any{"user@test.com", "plain text"},
	}
	result := RedactPII(data)
	items := result["items"].([]any)
	testkit.AssertNotEqual(t, items[0], "user@test.com")
	testkit.AssertEqual(t, items[1], "plain text")
}

func TestRedactPII_NonString(t *testing.T) {
	data := map[string]any{"count": 42}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["count"], 42)
}

func TestRedactPII_Empty(t *testing.T) {
	result := RedactPII(map[string]any{})
	testkit.AssertLen(t, result, 0)
}

func TestRedactPII_CaseInsensitiveField(t *testing.T) {
	data := map[string]any{"Password": "secret123"}
	result := RedactPII(data)
	// "Password" has capital P, but piiFields has "password" lowercase
	// The code checks both the original key and lowercase
	testkit.AssertEqual(t, result["Password"], "[REDACTED]")
}

func TestContainsSuspiciousInput_MixedCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"DrOp TaBlE users", true},
		{"SeLeCt * FrOm users", true},
		{"<ScRiPt>alert(1)</ScRiPt>", true},
		{"JaVaScRiPt: alert(1)", true},
		{"OnClIcK=doEvil()", true},
	}
	for _, tt := range tests {
		testkit.AssertEqual(t, ContainsSuspiciousInput(tt.input), tt.want)
	}
}

func TestRedactPII_DeeplyNested(t *testing.T) {
	data := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": map[string]any{
					"email": "deep@test.com",
					"safe":  "visible",
				},
			},
		},
	}
	result := RedactPII(data)
	l1 := result["level1"].(map[string]any)
	l2 := l1["level2"].(map[string]any)
	l3 := l2["level3"].(map[string]any)
	testkit.AssertEqual(t, l3["email"], "[REDACTED]")
	testkit.AssertEqual(t, l3["safe"], "visible")
}

func TestRedactPII_PhoneFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"dash", "Call 555-123-4567"},
		{"dot", "Call 555.123.4567"},
		{"parens", "Call (555) 123-4567"},
		{"plus1", "Call +1-555-123-4567"},
		{"space", "Call 555 123 4567"},
	}
	for _, tt := range tests {
		data := map[string]any{"msg": tt.input}
		result := RedactPII(data)
		got := result["msg"].(string)
		testkit.AssertNotEqual(t, got, tt.input)
	}
}

func TestRedactPII_BoolAndNilValues(t *testing.T) {
	data := map[string]any{
		"active": true,
		"score":  3.14,
		"empty":  nil,
	}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["active"], true)
	testkit.AssertEqual(t, result["score"], 3.14)
	testkit.AssertNil(t, result["empty"])
}

func TestRedactPII_ArrayOfMaps(t *testing.T) {
	data := map[string]any{
		"users": []any{
			map[string]any{"email": "a@test.com", "name": "A"},
			map[string]any{"email": "b@test.com", "name": "B"},
		},
	}
	result := RedactPII(data)
	users := result["users"].([]any)
	for _, u := range users {
		m := u.(map[string]any)
		testkit.AssertEqual(t, m["email"], "[REDACTED]")
	}
}

func TestContainsSuspiciousInput_SafeInputs(t *testing.T) {
	safe := []string{
		"Hello World",
		"This is a normal comment with <b>HTML</b>",
		"SELECT your favorite",
		"Drop your resume here",
		"Update your profile",
		"Delete this draft",
	}
	for _, s := range safe {
		testkit.AssertFalse(t, ContainsSuspiciousInput(s))
	}
}

func TestRedactPII_NilMap(t *testing.T) {
	result := RedactPII(nil)
	testkit.AssertNotNil(t, result)
	testkit.AssertLen(t, result, 0)
}

func TestRedactPII_AllPIIFieldNames(t *testing.T) {
	fields := []string{
		"email", "phone", "address", "ssn", "password", "resume",
		"socialSecurityNumber", "phoneNumber", "emailAddress",
		"streetAddress", "zipCode", "postalCode", "dateOfBirth",
	}
	for _, field := range fields {
		data := map[string]any{field: "sensitive-value"}
		result := RedactPII(data)
		testkit.AssertEqual(t, result[field], "[REDACTED]")
	}
}

func TestRedactPII_NonPIIFieldPassthrough(t *testing.T) {
	data := map[string]any{
		"status":    "active",
		"count":     int64(99),
		"threshold": 0.5,
		"flag":      false,
	}
	result := RedactPII(data)
	testkit.AssertEqual(t, result["status"], "active")
	testkit.AssertEqual(t, result["count"], int64(99))
	testkit.AssertEqual(t, result["threshold"], 0.5)
	testkit.AssertEqual(t, result["flag"], false)
}

func TestRedactPII_EmptyArray(t *testing.T) {
	data := map[string]any{"items": []any{}}
	result := RedactPII(data)
	items := result["items"].([]any)
	testkit.AssertLen(t, items, 0)
}

func TestRedactPII_MixedArrayTypes(t *testing.T) {
	data := map[string]any{
		"mixed": []any{"user@test.com", 42, true, nil, "clean text"},
	}
	result := RedactPII(data)
	items := result["mixed"].([]any)
	testkit.AssertNotEqual(t, items[0], "user@test.com")
	testkit.AssertEqual(t, items[1], 42)
	testkit.AssertEqual(t, items[2], true)
	testkit.AssertNil(t, items[3])
	testkit.AssertEqual(t, items[4], "clean text")
}

func TestContainsSuspiciousInput_AllPatterns(t *testing.T) {
	// Verify each suspicious pattern individually
	tests := []struct {
		name  string
		input string
	}{
		{"DROP TABLE", "DROP TABLE users"},
		{"SELECT * FROM", "SELECT * FROM accounts"},
		{"UNION SELECT", "1 UNION SELECT password"},
		{"INSERT INTO", "INSERT INTO logs VALUES"},
		{"DELETE FROM", "DELETE FROM sessions"},
		{"UPDATE SET", "UPDATE users SET role"},
		{"script tag", "<script>alert(1)</script>"},
		{"javascript:", "javascript: void(0)"},
		{"event handler", "onload=hack()"},
		{"iframe", "<iframe src=evil>"},
		{"object", "<object data=x>"},
		{"embed", "<embed src=x>"},
		{"svg event", "<svg onclick=x>"},
		{"data URI", "data: text/html,payload"},
	}
	for _, tt := range tests {
		testkit.AssertTrue(t, ContainsSuspiciousInput(tt.input))
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
