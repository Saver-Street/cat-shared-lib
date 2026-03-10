// Package flags provides feature flag lookups backed by a site_settings database table.
package flags

// Flag name constants — the key suffix used in site_settings (prefixed with "flag_").
const (
	// FlagAIScoring controls the AI-powered application scoring feature.
	FlagAIScoring = "aiScoring"
	// FlagResumeParsing controls automated resume parsing on upload.
	FlagResumeParsing = "resumeParsing"
	// FlagMaintenanceMode puts the application into read-only maintenance mode.
	FlagMaintenanceMode = "maintenanceMode"
	// FlagGlobalAutoPause pauses all automated application submissions globally.
	FlagGlobalAutoPause = "globalAutomationPause"
	// FlagGmailOAuth controls the Gmail OAuth integration feature.
	FlagGmailOAuth = "gmailOAuth"
	// FlagAuditLogging controls server-side audit trail logging.
	FlagAuditLogging = "auditLogging"
	// FlagSIAFI controls the SIAFI (Smart Intelligent Application Filing) mode.
	FlagSIAFI = "siafiMode"
	// FlagCoverLetterGen controls the AI cover letter generation feature.
	FlagCoverLetterGen = "coverLetterGen"
)
