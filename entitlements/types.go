// Package entitlements provides subscription tier lookups and usage checks.
package entitlements

// TierLimits defines the feature caps for a subscription tier.
type TierLimits struct {
	Tier                string `json:"tier"`
	MonthlyApplications int    `json:"monthlyApplications"`
	ResumeVersions      int    `json:"resumeVersions"`
	TracksLimit         int    `json:"tracksLimit"`
	AutoSubmit          bool   `json:"autoSubmit"`
	SIAFIMode           bool   `json:"siafiMode"`
	GmailIntegration    bool   `json:"gmailIntegration"`
	CoverLetterGen      bool   `json:"coverLetterGen"`
	ContinuousEmail     bool   `json:"continuousEmail"`
	LearningMemory      bool   `json:"learningMemory"`
	AdvancedAnalytics   bool   `json:"advancedAnalytics"`
	PriorityQueue       bool   `json:"priorityQueue"`
	CoachingSessions    int    `json:"coachingSessions"`
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
