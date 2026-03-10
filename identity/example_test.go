package identity_test

import (
	"fmt"
	"net/http"

	"github.com/Saver-Street/cat-shared-lib/identity"
	"github.com/Saver-Street/cat-shared-lib/middleware"
)

func ExampleGetUserID() {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	// Without a user ID in context, returns empty string.
	fmt.Printf("empty: %q\n", identity.GetUserID(r))

	// With user ID set via middleware.SetUserID (as done by JWT auth middleware).
	r = r.WithContext(middleware.SetUserID(r.Context(), "user-123"))
	fmt.Printf("with user: %q\n", identity.GetUserID(r))
	// Output:
	// empty: ""
	// with user: "user-123"
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
