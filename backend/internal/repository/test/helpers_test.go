package repository_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// Helper function to create test users
func createTestUser(t *testing.T, db *sql.DB, id uuid.UUID, username, email string) {
	t.Helper()
	query := `
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := db.Exec(query, id, username, email, "hash", time.Now(), time.Now())
	require.NoError(t, err)
}

// Helper function to create test conversation
func createTestConversation(t *testing.T, db *sql.DB, user1ID, user2ID uuid.UUID) uuid.UUID {
	t.Helper()
	convID := uuid.New()

	// Create conversation
	_, err := db.Exec(`
		INSERT INTO conversations (id, created_at, updated_at)
		VALUES ($1, $2, $3)
	`, convID, time.Now(), time.Now())
	require.NoError(t, err)

	// Add participants
	_, err = db.Exec(`
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES ($1, $2), ($1, $3)
	`, convID, user1ID, user2ID)
	require.NoError(t, err)

	return convID
}

// Helper function to create test group
func createTestGroup(t *testing.T, db *sql.DB, name string, createdBy uuid.UUID, members []uuid.UUID) uuid.UUID {
	t.Helper()
	groupID := uuid.New()

	// Create group
	_, err := db.Exec(`
		INSERT INTO groups (id, name, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, groupID, name, createdBy, time.Now(), time.Now())
	require.NoError(t, err)

	// Add members
	for _, memberID := range members {
		_, err = db.Exec(`
			INSERT INTO group_members (group_id, user_id)
			VALUES ($1, $2)
		`, groupID, memberID)
		require.NoError(t, err)
	}

	return groupID
}
