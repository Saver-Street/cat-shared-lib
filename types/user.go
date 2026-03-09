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

// CandidateProfile represents the job-seeker profile linked to a User.
type CandidateProfile struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}
