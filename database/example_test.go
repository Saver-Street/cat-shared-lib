package database_test

import (
	"fmt"
	"time"

	"github.com/Saver-Street/cat-shared-lib/database"
)

func ExamplePoolConfig() {
	cfg := database.PoolConfig{
		DSN:             "postgres://user:pass@localhost:5432/mydb",
		MaxConns:        20,
		MinConns:        4,
		MaxConnLifetime: 2 * time.Hour,
		MaxConnIdleTime: 15 * time.Minute,
	}
	fmt.Println(cfg.MaxConns)
	fmt.Println(cfg.MinConns)
	// Output:
	// 20
	// 4
}

func ExampleMigration() {
	m := database.Migration{
		Version: 1,
		Name:    "create_users",
		SQL:     "CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT NOT NULL)",
	}
	fmt.Printf("v%d: %s\n", m.Version, m.Name)
	// Output:
	// v1: create_users
}
