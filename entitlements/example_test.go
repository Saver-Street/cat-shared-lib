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

func ExampleAllTierNames() {
	names := entitlements.AllTierNames()
	fmt.Println(names)
	// Output:
	// [free starter pro power concierge]
}

func ExampleCanApplyThisMonth() {
	fmt.Println(entitlements.CanApplyThisMonth("free", 9))
	fmt.Println(entitlements.CanApplyThisMonth("free", 10))
	// Output:
	// true
	// false
}
