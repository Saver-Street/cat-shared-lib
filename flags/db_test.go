package flags

import (
	"context"
	"testing"
)

func TestIsFeatureEnabled_NilPool(t *testing.T) {
	if !IsFeatureEnabled(context.TODO(), nil, FlagAIScoring) {
		t.Error("nil pool should return true (safe default)")
	}
}

func TestIsMaintenanceModeActive_NilPool(t *testing.T) {
	if IsMaintenanceModeActive(context.TODO(), nil) {
		t.Error("nil pool should return false for maintenance mode")
	}
}

func TestIsGlobalAutomationPaused_NilPool(t *testing.T) {
	if IsGlobalAutomationPaused(context.TODO(), nil) {
		t.Error("nil pool should return false for automation pause")
	}
}

func TestFlagConstants_NonEmpty(t *testing.T) {
	constants := []string{
		FlagAIScoring,
		FlagResumeParsing,
		FlagMaintenanceMode,
		FlagGlobalAutoPause,
		FlagGmailOAuth,
		FlagAuditLogging,
		FlagSIAFI,
		FlagCoverLetterGen,
	}
	for _, c := range constants {
		if c == "" {
			t.Errorf("flag constant should not be empty")
		}
	}
}

func TestFlagConstants_Unique(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range []string{
		FlagAIScoring, FlagResumeParsing, FlagMaintenanceMode,
		FlagGlobalAutoPause, FlagGmailOAuth, FlagAuditLogging,
		FlagSIAFI, FlagCoverLetterGen,
	} {
		if seen[c] {
			t.Errorf("duplicate flag constant: %q", c)
		}
		seen[c] = true
	}
}
