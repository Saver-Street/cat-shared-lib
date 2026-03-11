// Package types defines shared domain types used across Catherine
// microservices, including pagination parameters, user accounts, and candidate
// profiles.
//
// [PaginationParams] carries limit/offset values with helpers like
// [PaginationParams.HasNextPage], [PaginationParams.TotalPages], and
// [NormalizePage] for safe page-to-offset conversion.
//
// [User] represents an authenticated account with subscription fields and
// convenience methods [User.IsAdmin], [User.IsActive], [User.IsTrialing], and
// [User.HasAccess].  [CandidateProfile] holds a job-seeker's profile linked
// to a [User], with [CandidateProfile.FullName] for display purposes.
package types
