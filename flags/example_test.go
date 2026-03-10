package flags_test

import (
	"context"
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/flags"
)

func ExampleIsFeatureEnabled_nilPool() {
	// With nil DB, feature flags default to enabled (safe default)
	enabled := flags.IsFeatureEnabled(context.Background(), nil, flags.FlagAIScoring)
	fmt.Println(enabled)
	// Output:
	// true
}

func ExampleIsMaintenanceModeActive_nilPool() {
	// With nil DB, maintenance mode defaults to inactive
	active := flags.IsMaintenanceModeActive(context.Background(), nil)
	fmt.Println(active)
	// Output:
	// false
}

func ExampleIsGlobalAutomationPaused_nilPool() {
	// With nil DB, automation defaults to running
	paused := flags.IsGlobalAutomationPaused(context.Background(), nil)
	fmt.Println(paused)
	// Output:
	// false
}
