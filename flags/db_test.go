package flags

import (
	"context"
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestIsFeatureEnabled_NilPool(t *testing.T) {
	testkit.AssertTrue(t, IsFeatureEnabled(context.TODO(), nil, FlagAIScoring))
}

func TestIsMaintenanceModeActive_NilPool(t *testing.T) {
	testkit.AssertFalse(t, IsMaintenanceModeActive(context.TODO(), nil))
}

func TestIsGlobalAutomationPaused_NilPool(t *testing.T) {
	testkit.AssertFalse(t, IsGlobalAutomationPaused(context.TODO(), nil))
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

func TestIsCustomFlagEnabled_NilDB_DefaultTrue(t *testing.T) {
	testkit.AssertTrue(t, IsCustomFlagEnabled(context.TODO(), nil, "any_key", true))
}

func TestIsCustomFlagEnabled_NilDB_DefaultFalse(t *testing.T) {
	testkit.AssertFalse(t, IsCustomFlagEnabled(context.TODO(), nil, "any_key", false))
}
