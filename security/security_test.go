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
		if got := ContainsSuspiciousInput(tt.input); got != tt.want {
			t.Errorf("ContainsSuspiciousInput(%q) = %v, want %v", tt.input, got, tt.want)
		}
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
	if l3["email"] != "[REDACTED]" {
		t.Errorf("deeply nested email not redacted: %v", l3["email"])
	}
	if l3["safe"] != "visible" {
		t.Errorf("safe field changed: %v", l3["safe"])
	}
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
		if got == tt.input {
			t.Errorf("%s: phone not redacted in %q", tt.name, tt.input)
		}
	}
}

func TestRedactPII_BoolAndNilValues(t *testing.T) {
	data := map[string]any{
		"active": true,
		"score":  3.14,
		"empty":  nil,
	}
	result := RedactPII(data)
	if result["active"] != true {
		t.Errorf("bool changed: %v", result["active"])
	}
	if result["score"] != 3.14 {
		t.Errorf("float changed: %v", result["score"])
	}
	if result["empty"] != nil {
		t.Errorf("nil changed: %v", result["empty"])
	}
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
	for i, u := range users {
		m := u.(map[string]any)
		if m["email"] != "[REDACTED]" {
			t.Errorf("users[%d].email not redacted: %v", i, m["email"])
		}
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
		if ContainsSuspiciousInput(s) {
			t.Errorf("safe input flagged: %q", s)
		}
	}
}

func TestRedactPII_NilMap(t *testing.T) {
	result := RedactPII(nil)
	if result == nil {
		t.Error("RedactPII(nil) should return non-nil empty map")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %v", result)
	}
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
		if result[field] != "[REDACTED]" {
			t.Errorf("field %q not redacted: got %v", field, result[field])
		}
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
	if result["status"] != "active" {
		t.Errorf("status changed: %v", result["status"])
	}
	if result["count"] != int64(99) {
		t.Errorf("count changed: %v", result["count"])
	}
	if result["threshold"] != 0.5 {
		t.Errorf("threshold changed: %v", result["threshold"])
	}
	if result["flag"] != false {
		t.Errorf("flag changed: %v", result["flag"])
	}
}

func TestRedactPII_EmptyArray(t *testing.T) {
	data := map[string]any{"items": []any{}}
	result := RedactPII(data)
	items := result["items"].([]any)
	if len(items) != 0 {
		t.Errorf("expected empty array, got %v", items)
	}
}

func TestRedactPII_MixedArrayTypes(t *testing.T) {
	data := map[string]any{
		"mixed": []any{"user@test.com", 42, true, nil, "clean text"},
	}
	result := RedactPII(data)
	items := result["mixed"].([]any)
	if items[0] == "user@test.com" {
		t.Error("email in array not redacted")
	}
	if items[1] != 42 {
		t.Errorf("int changed: %v", items[1])
	}
	if items[2] != true {
		t.Errorf("bool changed: %v", items[2])
	}
	if items[3] != nil {
		t.Errorf("nil changed: %v", items[3])
	}
	if items[4] != "clean text" {
		t.Errorf("clean text changed: %v", items[4])
	}
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
		if !ContainsSuspiciousInput(tt.input) {
			t.Errorf("pattern %q not detected in %q", tt.name, tt.input)
		}
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

func TestRedactPII_DeepNesting_Truncated(t *testing.T) {
// Build a map nested deeper than maxRedactDepth to exercise the truncation path.
var build func(depth int) map[string]any
build = func(depth int) map[string]any {
if depth == 0 {
return map[string]any{"leaf": "value"}
}
return map[string]any{"nested": build(depth - 1)}
}
deep := build(maxRedactDepth + 5)
result := RedactPII(deep)
// Traverse until we hit the truncation sentinel.
var traverse func(v any, depth int) bool
traverse = func(v any, depth int) bool {
if s, ok := v.(string); ok && s == "[TRUNCATED]" {
return true
}
if m, ok := v.(map[string]any); ok {
for _, child := range m {
if traverse(child, depth+1) {
return true
}
}
}
return false
}
if !traverse(result, 0) {
t.Error("expected [TRUNCATED] sentinel in deeply nested RedactPII output")
}
}
