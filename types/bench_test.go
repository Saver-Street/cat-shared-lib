package types

import "testing"

func BenchmarkHasNextPage(b *testing.B) {
	p := NormalizePage(2, 10)
	for b.Loop() {
		p.HasNextPage(50)
	}
}

func BenchmarkTotalPages(b *testing.B) {
	p := NormalizePage(1, 10)
	for b.Loop() {
		p.TotalPages(97)
	}
}

func BenchmarkUserIsAdmin(b *testing.B) {
	u := User{Role: "admin"}
	for b.Loop() {
		u.IsAdmin()
	}
}

func BenchmarkUserHasAccess(b *testing.B) {
	u := User{SubscriptionStatus: "active"}
	for b.Loop() {
		u.HasAccess()
	}
}

func BenchmarkCandidateProfile_FullName(b *testing.B) {
	c := CandidateProfile{FirstName: "Jane", LastName: "Doe"}
	for b.Loop() {
		c.FullName()
	}
}
