package migration

import "testing"

func BenchmarkValidateMigrations_Valid(b *testing.B) {
	migrations := []Migration{
		{ID: 1, Name: "create_users", Up: "CREATE TABLE users (id SERIAL PRIMARY KEY)", Down: "DROP TABLE users"},
		{ID: 2, Name: "create_orders", Up: "CREATE TABLE orders (id SERIAL PRIMARY KEY)", Down: "DROP TABLE orders"},
		{ID: 3, Name: "add_user_email", Up: "ALTER TABLE users ADD COLUMN email TEXT", Down: "ALTER TABLE users DROP COLUMN email"},
	}
	for b.Loop() {
		ValidateMigrations(migrations)
	}
}

func BenchmarkValidateMigrations_Large(b *testing.B) {
	migrations := make([]Migration, 50)
	for i := range migrations {
		migrations[i] = Migration{
			ID:   i + 1,
			Name: "migration",
			Up:   "SELECT 1",
			Down: "SELECT 1",
		}
	}
	for b.Loop() {
		ValidateMigrations(migrations)
	}
}

func BenchmarkValidateMigrations_WithErrors(b *testing.B) {
	migrations := []Migration{
		{ID: 0, Name: "", Up: "", Down: ""},
		{ID: -1, Name: "", Up: "", Down: ""},
		{ID: 1, Name: "valid", Up: "SELECT 1", Down: ""},
		{ID: 1, Name: "duplicate", Up: "SELECT 1", Down: "SELECT 1"},
	}
	for b.Loop() {
		ValidateMigrations(migrations)
	}
}

func BenchmarkValidationError_Error(b *testing.B) {
	migrations := []Migration{
		{ID: 0, Name: "", Up: "", Down: ""},
		{ID: -1, Name: "", Up: "", Down: ""},
	}
	err := ValidateMigrations(migrations)
	for b.Loop() {
		_ = err.Error()
	}
}
