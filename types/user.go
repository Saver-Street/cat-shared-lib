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
func (u User) IsActive() bool { return u.SubscriptionStatus == "active" }

// CandidateProfile represents the job-seeker profile linked to a User.
type CandidateProfile struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
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
