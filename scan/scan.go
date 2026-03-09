// Package scan provides generic database row scanning helpers.
package scan

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
func Rows[T any](rows RowScanner, scanFn func(*T) []any) ([]T, error) {
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
func Row[T any](row SingleRowScanner, scanFn func(*T) []any) (T, error) {
	var item T
	if err := row.Scan(scanFn(&item)...); err != nil {
		return item, err
	}
	return item, nil
}
