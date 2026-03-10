// Package scan provides generic database row scanning helpers.
package scan

import "errors"

// RowScanner is satisfied by both *sql.Rows and pgx.Rows.
type RowScanner interface {
	Next() bool
	Scan(dest ...any) error
	Close()
	Err() error
}

// SingleRowScanner is satisfied by *sql.Row and pgx.Row.
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

	results := make([]T, 0)
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
