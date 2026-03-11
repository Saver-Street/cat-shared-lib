package migration_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/migration"
)

func ExampleDefaultTable() {
	fmt.Println(migration.DefaultTable)
	// Output:
	// schema_migrations
}

func ExampleMigration() {
	m := migration.Migration{
		ID:   1,
		Name: "create_users",
		Up:   "CREATE TABLE users (id SERIAL PRIMARY KEY, email TEXT NOT NULL)",
		Down: "DROP TABLE users",
	}
	fmt.Printf("migration %d: %s\n", m.ID, m.Name)
	// Output:
	// migration 1: create_users
}

func ExampleErrDuplicateID() {
	fmt.Println(migration.ErrDuplicateID)
	// Output:
	// migration: duplicate migration ID
}

func ExampleValidateMigrations() {
	migrations := []migration.Migration{
		{ID: 1, Name: "create_users", Up: "CREATE TABLE users (id INT)", Down: "DROP TABLE users"},
		{ID: 2, Name: "create_posts", Up: "CREATE TABLE posts (id INT)", Down: "DROP TABLE posts"},
	}
	if err := migration.ValidateMigrations(migrations); err != nil {
		fmt.Println("validation failed:", err)
		return
	}
	fmt.Println("all migrations valid")
	// Output:
	// all migrations valid
}

func ExampleValidateMigrations_errors() {
	migrations := []migration.Migration{
		{ID: 0, Name: "", Up: "", Down: ""},
	}
	err := migration.ValidateMigrations(migrations)
	if err != nil {
		fmt.Println("validation failed")
	}
	// Output:
	// validation failed
}
