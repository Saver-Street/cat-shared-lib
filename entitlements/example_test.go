package entitlements_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/entitlements"
)

func ExampleGetLimitsForTier() {
	limits := entitlements.GetLimitsForTier("pro")
	fmt.Printf("Tier=%s Apps=%d AutoSubmit=%v\n", limits.Tier, limits.MonthlyApplications, limits.AutoSubmit)
	// Output:
	// Tier=pro Apps=150 AutoSubmit=true
}

func ExampleGetLimitsForTier_unknown() {
	limits := entitlements.GetLimitsForTier("enterprise")
	fmt.Println(limits.Tier)
	// Output:
	// free
}
