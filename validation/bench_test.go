package validation

import "testing"

func BenchmarkEmail(b *testing.B) {
	for b.Loop() {
		Email("email", "user@example.com")
	}
}

func BenchmarkUUID(b *testing.B) {
	for b.Loop() {
		UUID("id", "550e8400-e29b-41d4-a716-446655440000")
	}
}

func BenchmarkPhone(b *testing.B) {
	for b.Loop() {
		Phone("phone", "+1-555-123-4567")
	}
}

func BenchmarkURL(b *testing.B) {
	for b.Loop() {
		URL("website", "https://example.com/path?q=1")
	}
}

func BenchmarkRequired(b *testing.B) {
	for b.Loop() {
		Required("name", "Alice")
	}
}

func BenchmarkMinLength(b *testing.B) {
	for b.Loop() {
		MinLength("password", "supersecretpassword", 8)
	}
}

func BenchmarkCollect(b *testing.B) {
	errs := []error{
		Email("email", "bad"),
		Required("name", ""),
		nil,
		MinLength("pw", "short", 12),
	}
	for b.Loop() {
		Collect(errs...)
	}
}
