// Package testkit provides shared assertion utilities, HTTP test helpers, and
// mock infrastructure for Catherine service test suites.
//
// Assertion functions such as [AssertEqual], [AssertNoError], [AssertJSON],
// [AssertJSONContains], and [AssertContains] produce clear failure messages
// with minimal boilerplate.  HTTP helpers [NewRequest], [NewJSONRequest],
// [AssertStatus], and [AssertHeader] simplify handler testing.
//
// [NewMockServer] creates a [MockServer] backed by httptest.Server with
// request recording via [MockServer.RequestCount] and [MockServer.LastRequest].
// [CallRecorder] captures arbitrary function calls for later inspection.
//
// [ContextWithValue] and [MustMarshalJSON] round out the toolkit with
// context and serialization shortcuts.
package testkit
