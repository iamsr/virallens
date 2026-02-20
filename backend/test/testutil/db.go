package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// SetupTestDB creates a test database connection
func SetupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Fatal("TEST_DATABASE_URL environment variable not set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	return db
}

// CleanupTestDB cleans all tables in the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{
		"messages",
		"group_members",
		"groups",
		"conversation_participants",
		"conversations",
		"refresh_tokens",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}

// TeardownTestDB closes the database connection
func TeardownTestDB(t *testing.T, db *sql.DB) {
	t.Helper()
	if err := db.Close(); err != nil {
		t.Logf("warning: failed to close database: %v", err)
	}
}
