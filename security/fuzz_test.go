package security

import "testing"

func FuzzContainsSuspiciousInput(f *testing.F) {
	// Seed corpus with known patterns
	f.Add("normal text")
	f.Add("DROP TABLE users")
	f.Add("<script>alert(1)</script>")
	f.Add("SELECT * FROM users")
	f.Add("UNION SELECT password FROM users")
	f.Add("INSERT INTO users VALUES")
	f.Add("DELETE FROM users WHERE 1=1")
	f.Add("UPDATE users SET admin = true")
	f.Add("javascript: alert(1)")
	f.Add("onclick=alert(1)")
	f.Add("<iframe src=evil>")
	f.Add("<object data=evil>")
	f.Add("<embed src=evil>")
	f.Add("<svg onload=alert(1)>")
	f.Add("data: text/html,<h1>Hi</h1>")
	f.Add("")
	f.Add("   ")
	f.Add("Hello, World!")

	f.Fuzz(func(t *testing.T, input string) {
		// Should not panic on any input
		ContainsSuspiciousInput(input)
	})
}

func FuzzRedactPII(f *testing.F) {
	f.Add("user@example.com")
	f.Add("555-123-4567")
	f.Add("123-45-6789")
	f.Add("normal text with no PII")
	f.Add("")
	f.Add("mixed: email user@test.com and phone 555-111-2222")
	f.Add("SSN is 000-00-0000")

	f.Fuzz(func(t *testing.T, input string) {
		data := map[string]any{"field": input}
		result := RedactPII(data)
		// Should not panic and result should always have the key
		if _, ok := result["field"]; !ok {
			t.Error("redacted result missing 'field' key")
		}
	})
}
