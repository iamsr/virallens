package repository_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yourusername/virallens/backend/internal/domain"
	"github.com/yourusername/virallens/backend/internal/repository"
)

func TestGroupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	member1ID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, member1ID, "member1", "member1@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, member1ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Verify group was created
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, found.ID)
	assert.Equal(t, group.Name, found.Name)
	assert.Equal(t, group.CreatedBy, found.CreatedBy)
	assert.ElementsMatch(t, group.Members, found.Members)
}

func TestGroupRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test user
	creatorID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Equal(t, group.ID, found.ID)
	assert.Equal(t, group.Name, found.Name)
	assert.Equal(t, group.CreatedBy, found.CreatedBy)
}

func TestGroupRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	_, err := repo.GetByID(uuid.New())
	assert.ErrorIs(t, err, domain.ErrGroupNotFound)
}

func TestGroupRepository_ListByUserID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test users
	user1ID := uuid.New()
	user2ID := uuid.New()
	user3ID := uuid.New()
	createTestUser(t, db, user1ID, "user1", "user1@example.com")
	createTestUser(t, db, user2ID, "user2", "user2@example.com")
	createTestUser(t, db, user3ID, "user3", "user3@example.com")

	// Create groups
	group1 := &domain.Group{
		ID:        uuid.New(),
		Name:      "Group 1",
		CreatedBy: user1ID,
		Members:   []uuid.UUID{user1ID, user2ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	group2 := &domain.Group{
		ID:        uuid.New(),
		Name:      "Group 2",
		CreatedBy: user2ID,
		Members:   []uuid.UUID{user1ID, user2ID, user3ID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	require.NoError(t, repo.Create(group1))
	require.NoError(t, repo.Create(group2))

	// User1 should be in 2 groups
	groups, err := repo.ListByUserID(user1ID)
	require.NoError(t, err)
	assert.Len(t, groups, 2)

	// User3 should be in 1 group
	groups, err = repo.ListByUserID(user3ID)
	require.NoError(t, err)
	assert.Len(t, groups, 1)
	assert.Equal(t, group2.ID, groups[0].ID)
}

func TestGroupRepository_AddMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	newMemberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, newMemberID, "newmember", "newmember@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Add new member
	err = repo.AddMember(group.ID, newMemberID)
	require.NoError(t, err)

	// Verify member was added
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Len(t, found.Members, 2)
	assert.Contains(t, found.Members, newMemberID)
}

func TestGroupRepository_RemoveMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	memberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, memberID, "member", "member@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Remove member
	err = repo.RemoveMember(group.ID, memberID)
	require.NoError(t, err)

	// Verify member was removed
	found, err := repo.GetByID(group.ID)
	require.NoError(t, err)
	assert.Len(t, found.Members, 1)
	assert.NotContains(t, found.Members, memberID)
}

func TestGroupRepository_IsMember(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := repository.NewGroupRepository(db)

	// Create test users
	creatorID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()
	createTestUser(t, db, creatorID, "creator", "creator@example.com")
	createTestUser(t, db, memberID, "member", "member@example.com")
	createTestUser(t, db, nonMemberID, "nonmember", "nonmember@example.com")

	group := &domain.Group{
		ID:        uuid.New(),
		Name:      "Test Group",
		CreatedBy: creatorID,
		Members:   []uuid.UUID{creatorID, memberID},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(group)
	require.NoError(t, err)

	// Check if creator is member
	isMember, err := repo.IsMember(group.ID, creatorID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Check if member is member
	isMember, err = repo.IsMember(group.ID, memberID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Check if non-member is member
	isMember, err = repo.IsMember(group.ID, nonMemberID)
	require.NoError(t, err)
	assert.False(t, isMember)
}
