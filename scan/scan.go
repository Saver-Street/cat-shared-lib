// Package scan provides generic database row scanning helpers.
package scan

import "errors"

// RowScanner is the interface for iterating and scanning multiple query rows.
// It is satisfied by both *sql.Rows and pgx.Rows.
type RowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

// SingleRowScanner is the interface for scanning a single query row.
// It is satisfied by both *sql.Row and pgx.Row.
type SingleRowScanner interface {
	Scan(dest ...any) error
}

// Rows iterates over rows and scans each row into a value of type T.
// The scanFn returns a slice of pointers for Scan to populate.
// Returns an error if rows or scanFn is nil.
func Rows[T any](rows RowScanner, scanFn func(*T) []any) ([]T, error) {
	if rows == nil {
		return nil, errors.New("scan.Rows: rows must not be nil")
	}
	if scanFn == nil {
		return nil, errors.New("scan.Rows: scanFn must not be nil")
	}
	defer rows.Close()

	var results []T
	for rows.Next() {
		var item T
		if err := rows.Scan(scanFn(&item)...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// Row scans a single row into a value of type T.
// Returns an error if row or scanFn is nil.
func Row[T any](row SingleRowScanner, scanFn func(*T) []any) (T, error) {
	var item T
	if row == nil {
		return item, errors.New("scan.Row: row must not be nil")
	}
	if scanFn == nil {
		return item, errors.New("scan.Row: scanFn must not be nil")
	}
	if err := row.Scan(scanFn(&item)...); err != nil {
		return item, err
	}
	return item, nil
}

// First returns the first row from rows as a value of type T.
// It returns an error if rows or scanFn is nil, or if the result set is empty.
// The remaining rows, if any, are discarded.
func First[T any](rows RowScanner, scanFn func(*T) []any) (T, error) {
	var item T
	if rows == nil {
		return item, errors.New("scan.First: rows must not be nil")
	}
	if scanFn == nil {
		return item, errors.New("scan.First: scanFn must not be nil")
	}
	defer rows.Close()

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return item, err
		}
		return item, errors.New("scan.First: no rows in result set")
	}
	if err := rows.Scan(scanFn(&item)...); err != nil {
		return item, err
	}
	return item, nil
}

// RowsLimit is like Rows but stops scanning after at most limit rows.
// This prevents accidental large allocations when the caller knows the
// maximum useful result count (e.g., auto-complete, preview lists).
// If limit is zero or negative, all rows are returned.
func RowsLimit[T any](rows RowScanner, scanFn func(*T) []any, limit int) ([]T, error) {
	if rows == nil {
		return nil, errors.New("scan.RowsLimit: rows must not be nil")
	}
	if scanFn == nil {
		return nil, errors.New("scan.RowsLimit: scanFn must not be nil")
	}
	defer rows.Close()

	cap := limit
	if cap <= 0 {
		cap = 0
	}
	results := make([]T, 0, cap)
	for rows.Next() {
		if limit > 0 && len(results) >= limit {
			break
		}
		var item T
		if err := rows.Scan(scanFn(&item)...); err != nil {
			return nil, err
		}
		results = append(results, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
