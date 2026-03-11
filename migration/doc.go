// Package migration provides a lightweight database migration runner built on
// pgx/v5 that tracks applied migrations in a dedicated table.
//
// Create a [Runner] with [New], [NewWithTable] (custom tracking table name),
// or [NewWithDB] (custom [DB] interface).  Call [Runner.Init] to create the
// tracking table, then [Runner.Migrate] to apply all pending [Migration]
// entries in order.  Each migration is applied exactly once; duplicate IDs
// cause [ErrDuplicateID].
//
// [Runner.Rollback] reverses the most recently applied migration,
// [Runner.Applied] returns the list of applied [Record] entries, and
// [Runner.Status] reports which migrations are pending or applied.
// [DefaultTable] holds the default tracking table name ("schema_migrations").
package migration
