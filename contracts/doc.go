// Package contracts defines shared Go interfaces and types that all Catherine
// microservices must implement, enabling compile-time contract enforcement.
//
// The [Service] interface composes [ServiceHealth], [ServiceInfo], and
// [Handler].  Implementing [Service] ensures a microservice exposes a health
// check, basic metadata (name, version, environment), and an HTTP route
// registration method.
//
// [HealthStatus] and the [HealthState] constants ([HealthStateOK],
// [HealthStateDegraded], [HealthStateDown]) provide a standard vocabulary for
// health reporting.  [StandardError] and its constructors offer a uniform
// error envelope for JSON API responses.
package contracts
