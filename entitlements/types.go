// Package entitlements provides subscription tier lookups and usage checks.
package entitlements

// TierLimits defines the feature caps for a subscription tier.
type TierLimits struct {
	// Tier is the canonical tier name (e.g. "free", "starter", "pro").
	Tier string `json:"tier"`
	// MonthlyApplications is the maximum job applications per calendar month.
	MonthlyApplications int `json:"monthlyApplications"`
	// ResumeVersions is the number of saved resume versions allowed.
	ResumeVersions int `json:"resumeVersions"`
	// TracksLimit is the maximum number of active application tracks.
	TracksLimit int `json:"tracksLimit"`
	// AutoSubmit enables automated job application submission.
	AutoSubmit bool `json:"autoSubmit"`
	// SIAFIMode enables the Smart Intelligent Application Filing feature.
	SIAFIMode bool `json:"siafiMode"`
	// GmailIntegration enables Gmail OAuth for application tracking.
	GmailIntegration bool `json:"gmailIntegration"`
	// CoverLetterGen enables AI-generated cover letters.
	CoverLetterGen bool `json:"coverLetterGen"`
	// ContinuousEmail enables continuous email monitoring for job replies.
	ContinuousEmail bool `json:"continuousEmail"`
	// LearningMemory enables persistent AI learning from past applications.
	LearningMemory bool `json:"learningMemory"`
	// AdvancedAnalytics enables detailed application performance analytics.
	AdvancedAnalytics bool `json:"advancedAnalytics"`
	// PriorityQueue gives the user priority placement in the application queue.
	PriorityQueue bool `json:"priorityQueue"`
	// CoachingSessions is the number of human coaching sessions included per month.
	CoachingSessions int `json:"coachingSessions"`
}

// TierConfig maps subscription tier names to their corresponding feature limits and entitlements.
var TierConfig = map[string]TierLimits{
	"free": {
		Tier:                "free",
		MonthlyApplications: 10,
		ResumeVersions:      1,
		TracksLimit:         1,
		AutoSubmit:          false,
		SIAFIMode:           false,
		GmailIntegration:    false,
		CoverLetterGen:      false,
		ContinuousEmail:     false,
		LearningMemory:      false,
		AdvancedAnalytics:   false,
		PriorityQueue:       false,
		CoachingSessions:    0,
	},
	"starter": {
		Tier:                "starter",
		MonthlyApplications: 40,
		ResumeVersions:      1,
		TracksLimit:         3,
		AutoSubmit:          false,
		SIAFIMode:           false,
		GmailIntegration:    true,
		CoverLetterGen:      false,
		ContinuousEmail:     false,
		LearningMemory:      false,
		AdvancedAnalytics:   false,
		PriorityQueue:       false,
		CoachingSessions:    0,
	},
	"pro": {
		Tier:                "pro",
		MonthlyApplications: 150,
		ResumeVersions:      5,
		TracksLimit:         10,
		AutoSubmit:          true,
		SIAFIMode:           true,
		GmailIntegration:    true,
		CoverLetterGen:      true,
		ContinuousEmail:     true,
		LearningMemory:      true,
		AdvancedAnalytics:   false,
		PriorityQueue:       false,
		CoachingSessions:    0,
	},
	"power": {
		Tier:                "power",
		MonthlyApplications: 400,
		ResumeVersions:      10,
		TracksLimit:         20,
		AutoSubmit:          true,
		SIAFIMode:           true,
		GmailIntegration:    true,
		CoverLetterGen:      true,
		ContinuousEmail:     true,
		LearningMemory:      true,
		AdvancedAnalytics:   true,
		PriorityQueue:       true,
		CoachingSessions:    0,
	},
	"concierge": {
		Tier:                "concierge",
		MonthlyApplications: 400,
		ResumeVersions:      10,
		TracksLimit:         20,
		AutoSubmit:          true,
		SIAFIMode:           true,
		GmailIntegration:    true,
		CoverLetterGen:      true,
		ContinuousEmail:     true,
		LearningMemory:      true,
		AdvancedAnalytics:   true,
		PriorityQueue:       true,
		CoachingSessions:    1,
	},
}
