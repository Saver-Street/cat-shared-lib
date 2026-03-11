// Package openapi provides a fluent builder API for programmatically
// constructing OpenAPI 3.0 specification documents in Go.
//
// Start with [NewSpec] to create a [Spec], then chain methods like
// [Spec.AddServer] and [Spec.AddPath] to define your API surface.  Build
// operations with [NewOperation] and attach parameters, request bodies, and
// responses using the chainable methods on [Operation].
//
// Schema helpers [StringSchema], [IntegerSchema], [BooleanSchema],
// [ArraySchema], [ObjectSchema], and [RefSchema] simplify JSON Schema
// construction.
//
// Call [Spec.Handler] to obtain an http.HandlerFunc that serves the spec as
// JSON, or use [Spec.JSON] / [Spec.JSONIndent] to serialize it directly.
package openapi
