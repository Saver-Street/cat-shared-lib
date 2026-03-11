// Package servicetest provides integration test helpers for Catherine
// microservices, including HTTP test servers, database query mocks, and
// fixture management.
//
// [NewHTTPTestServer] creates an [HTTPTestServer] that records requests and
// supports per-route handler registration via [HTTPTestServer.Handle] and
// [HTTPTestServer.HandleJSON].  Use [HTTPTestServer.Requests] and
// [HTTPTestServer.LastRequest] to inspect captured traffic.
//
// [DBTestHelper] implements the Querier interface with a [MockRow] queue,
// enabling deterministic database testing without a live connection.
// [DBTestHelper.Queries] returns all recorded queries for assertion.
//
// [Fixtures] manages named test data blobs via [Fixtures.Register] and
// [Fixtures.Load], with JSON serialization support through
// [Fixtures.RegisterJSON] and [Fixtures.LoadInto].
package servicetest
