// Package scan provides generic database row scanning helpers that work with
// both database/sql and pgx row interfaces.
//
// [Rows] iterates a multi-row result set and returns a typed slice, using a
// caller-supplied function that maps struct fields to scan destinations.
// [Row] scans a single row, [First] returns only the first row from a
// multi-row result, and [RowsLimit] caps the number of rows scanned.
//
// The [RowScanner] and [SingleRowScanner] interfaces abstract over concrete
// driver types so the same scan functions work with sql.Rows, pgx.Rows,
// sql.Row, and pgx.Row.  [ErrNoRows] is returned when a result set is empty.
package scan
