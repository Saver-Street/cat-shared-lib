// Package identity bridges JWT-authenticated user IDs to application-level
// candidate profiles by querying the database.
//
// [GetUserID] and [GetExtCandidateID] extract the authenticated user's ID and
// optional extension-provided candidate ID from the request context (set by
// the middleware package).  [LookupCandidateID] queries the database for the
// candidate profile associated with a user ID, while [ResolveCandidate]
// combines both approaches, preferring an extension-provided ID when present.
//
// All database queries accept a [Querier] interface so they work with any
// pgx-compatible executor (pool, connection, or transaction).
package identity
