// Package database provides PostgreSQL connection pooling, transaction
// management, and schema migration helpers built on top of pgx/v5.
//
// Call [NewPool] with a [PoolConfig] to create a connection pool with sensible
// defaults for pool size, timeouts, and health checks.  Use [WithTx] to
// execute a [TxFunc] inside an automatically managed transaction that commits
// on success, rolls back on error, and recovers from panics.
//
// The [Migrate] function applies a slice of [Migration] entries in version
// order, creating a schema_version tracking table if it does not already exist.
// The [Querier] interface abstracts query execution so the same code can run
// against a pool, a connection, or a transaction.
package database
