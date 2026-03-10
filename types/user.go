// Package types provides shared domain types used across cat microservices.
package types

import "time"

// User represents an authenticated user account.
type User struct {
	ID                 string    `json:"id"`
	Email              string    `json:"email"`
	Role               string    `json:"role"`
	SubscriptionTier   string    `json:"subscriptionTier"`
	SubscriptionStatus string    `json:"subscriptionStatus"`
	CreatedAt          time.Time `json:"createdAt"`
}

// IsAdmin reports whether the user holds the "admin" role.
func (u User) IsAdmin() bool { return u.Role == "admin" }

// IsActive reports whether the user has an active subscription.
// Only "active" is considered active; use HasAccess to include trialing users.
func (u User) IsActive() bool { return u.SubscriptionStatus == "active" }

// IsTrialing reports whether the user is on a free trial.
func (u User) IsTrialing() bool { return u.SubscriptionStatus == "trialing" }

// HasAccess reports whether the user may use subscription features.
// Both "active" and "trialing" statuses grant access.
func (u User) HasAccess() bool { return u.IsActive() || u.IsTrialing() }

// CandidateProfile represents the job-seeker profile linked to a User.
type CandidateProfile struct {
	// ID is the UUID primary key of the candidate profile.
	ID string `json:"id"`
	// UserID is the foreign key referencing the owning User.
	UserID string `json:"userId"`
	// FirstName is the candidate's given name.
	FirstName string `json:"firstName"`
	// LastName is the candidate's family name.
	LastName string `json:"lastName"`
	// Email is the candidate's contact email (may differ from the login email).
	Email string `json:"email"`
	// CreatedAt is the UTC timestamp when the profile was created.
	CreatedAt time.Time `json:"createdAt"`
}

// FullName returns the candidate's full display name.
// If only one of FirstName or LastName is set, it is returned alone.
func (c CandidateProfile) FullName() string {
	switch {
	case c.FirstName == "":
		return c.LastName
	case c.LastName == "":
		return c.FirstName
	default:
		return c.FirstName + " " + c.LastName
	}
}
