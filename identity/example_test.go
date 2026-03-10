package identity_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/identity"
)

func ExampleGetUserID() {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Without context value, returns empty string.
	fmt.Printf("empty: %q\n", identity.GetUserID(r))

	// With context value set (typically by auth middleware).
	type ctxKey string
	// Simulate what middleware does internally — set via context.
	ctx := context.WithValue(r.Context(), ctxKey("userId"), "user-123")
	r = r.WithContext(ctx)
	// Note: identity.GetUserID uses its own unexported key, so direct
	// context.WithValue with a different key type won't match.
	// This example demonstrates the empty-return behavior.
	fmt.Printf("wrong key type: %q\n", identity.GetUserID(r))
	// Output:
	// empty: ""
	// wrong key type: ""
}

func ExampleGetExtCandidateID() {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	fmt.Printf("empty: %q\n", identity.GetExtCandidateID(r))
	// Output:
	// empty: ""
}

func ExampleResolveCandidate_noIdentity() {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	id, err := identity.ResolveCandidate(r, nil)
	fmt.Printf("id=%q err=%v\n", id, err)
	// Output:
	// id="" err=<nil>
}
