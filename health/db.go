package health

import "context"

// Pinger is satisfied by *pgxpool.Pool and similar database pools.
type Pinger interface {
	Ping(ctx context.Context) error
}

// DBChecker returns a Checker that pings the database pool.
func DBChecker(pool Pinger) Checker {
	return NewChecker("db", func(ctx context.Context) error {
		return pool.Ping(ctx)
	})
}
