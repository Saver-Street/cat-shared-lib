package flags

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func FuzzIsFeatureEnabled(f *testing.F) {
	f.Add("ai_scoring")
	f.Add("resume_parsing")
	f.Add("")
	f.Add("flag with spaces")
	f.Add("flag_with_underscore")
	f.Add("a")
	f.Add("SELECT 1; DROP TABLE users;--")

	f.Fuzz(func(t *testing.T, flagName string) {
		db := &mockQuerier{queryRowFunc: func(_ context.Context, _ string, _ ...any) pgx.Row {
			return newValueRow("true")
		}}
		// Must not panic
		IsFeatureEnabled(context.Background(), db, flagName)
	})
}
