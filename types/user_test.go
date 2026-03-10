package types

import (
	"encoding/json"
	"testing"
	"time"
)

func TestUser_JSONRoundTrip(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	u := User{
		ID:                 "usr-123",
		Email:              "alice@example.com",
		Role:               "admin",
		SubscriptionTier:   "pro",
		SubscriptionStatus: "active",
		CreatedAt:          now,
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got User
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != u {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", got, u)
	}
}

func TestUser_JSONFieldNames(t *testing.T) {
	u := User{
		ID:                 "u1",
		SubscriptionTier:   "free",
		SubscriptionStatus: "active",
	}
	data, err := json.Marshal(u)
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"id", "email", "role", "subscriptionTier", "subscriptionStatus", "createdAt"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON missing expected key %q", key)
		}
	}
}

func TestUser_ZeroValue(t *testing.T) {
	var u User
	if u.ID != "" || u.Email != "" || u.Role != "" {
		t.Error("zero-value User should have empty string fields")
	}
	if !u.CreatedAt.IsZero() {
		t.Error("zero-value CreatedAt should be zero time")
	}
}

func TestCandidateProfile_JSONRoundTrip(t *testing.T) {
	now := time.Date(2025, 3, 1, 9, 30, 0, 0, time.UTC)
	cp := CandidateProfile{
		ID:        "cand-456",
		UserID:    "usr-123",
		FirstName: "Alice",
		LastName:  "Smith",
		Email:     "alice@example.com",
		CreatedAt: now,
	}
	data, err := json.Marshal(cp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got CandidateProfile
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != cp {
		t.Errorf("round trip mismatch:\n got %+v\nwant %+v", got, cp)
	}
}

func TestCandidateProfile_JSONFieldNames(t *testing.T) {
	cp := CandidateProfile{ID: "c1", UserID: "u1"}
	data, _ := json.Marshal(cp)

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"id", "userId", "firstName", "lastName", "email", "createdAt"} {
		if _, ok := m[key]; !ok {
			t.Errorf("JSON missing expected key %q", key)
		}
	}
}

func TestCandidateProfile_ZeroValue(t *testing.T) {
	var cp CandidateProfile
	if cp.ID != "" || cp.UserID != "" || cp.FirstName != "" {
		t.Error("zero-value CandidateProfile should have empty string fields")
	}
}

func BenchmarkUserJSON(b *testing.B) {
	u := User{
		ID:                 "usr-bench",
		Email:              "bench@example.com",
		Role:               "user",
		SubscriptionTier:   "pro",
		SubscriptionStatus: "active",
		CreatedAt:          time.Now(),
	}
	for b.Loop() {
		data, _ := json.Marshal(u)
		var out User
		json.Unmarshal(data, &out)
	}
}

func BenchmarkCandidateProfileJSON(b *testing.B) {
	cp := CandidateProfile{
		ID:        "cand-bench",
		UserID:    "usr-bench",
		FirstName: "Bench",
		LastName:  "Mark",
		Email:     "bench@example.com",
		CreatedAt: time.Now(),
	}
	for b.Loop() {
		data, _ := json.Marshal(cp)
		var out CandidateProfile
		json.Unmarshal(data, &out)
	}
}

func TestUser_IsAdmin(t *testing.T) {
	admin := User{Role: "admin"}
	user := User{Role: "user"}
	anon := User{}
	if !admin.IsAdmin() {
		t.Error("IsAdmin() should return true for role=admin")
	}
	if user.IsAdmin() {
		t.Error("IsAdmin() should return false for role=user")
	}
	if anon.IsAdmin() {
		t.Error("IsAdmin() should return false for empty role")
	}
}

func TestUser_IsActive(t *testing.T) {
	active := User{SubscriptionStatus: "active"}
	inactive := User{SubscriptionStatus: "past_due"}
	zero := User{}
	if !active.IsActive() {
		t.Error("IsActive() should return true for status=active")
	}
	if inactive.IsActive() {
		t.Error("IsActive() should return false for status=past_due")
	}
	if zero.IsActive() {
		t.Error("IsActive() should return false for empty status")
	}
}

func TestCandidateProfile_FullName(t *testing.T) {
	tests := []struct {
		first, last, want string
	}{
		{"Alice", "Smith", "Alice Smith"},
		{"Alice", "", "Alice"},
		{"", "Smith", "Smith"},
		{"", "", ""},
	}
	for _, tc := range tests {
		cp := CandidateProfile{FirstName: tc.first, LastName: tc.last}
		if got := cp.FullName(); got != tc.want {
			t.Errorf("FullName(%q, %q) = %q, want %q", tc.first, tc.last, got, tc.want)
		}
	}
}

func TestUser_IsTrialing(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{"trialing", true},
		{"active", false},
		{"past_due", false},
		{"", false},
	}
	for _, tc := range cases {
		u := User{SubscriptionStatus: tc.status}
		if got := u.IsTrialing(); got != tc.want {
			t.Errorf("IsTrialing(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestUser_HasAccess(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{"active", true},
		{"trialing", true},
		{"past_due", false},
		{"canceled", false},
		{"", false},
	}
	for _, tc := range cases {
		u := User{SubscriptionStatus: tc.status}
		if got := u.HasAccess(); got != tc.want {
			t.Errorf("HasAccess(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}
