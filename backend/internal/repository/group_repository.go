package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

type groupRepository struct {
	db *sql.DB
}

// NewGroupRepository creates a new group repository
func NewGroupRepository(db *sql.DB) domain.GroupRepository {
	return &groupRepository{db: db}
}

func (r *groupRepository) Create(group *domain.Group) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert group
	query := `
		INSERT INTO groups (id, name, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(query, group.ID, group.Name, group.CreatedBy, group.CreatedAt, group.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert members
	for _, memberID := range group.Members {
		memberQuery := `
			INSERT INTO group_members (group_id, user_id)
			VALUES ($1, $2)
		`
		_, err = tx.Exec(memberQuery, group.ID, memberID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *groupRepository) GetByID(id uuid.UUID) (*domain.Group, error) {
	// Get group
	query := `
		SELECT id, name, created_by, created_at, updated_at
		FROM groups
		WHERE id = $1
	`

	group := &domain.Group{}
	err := r.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedBy,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrGroupNotFound
		}
		return nil, err
	}

	// Get members
	membersQuery := `
		SELECT user_id
		FROM group_members
		WHERE group_id = $1
	`

	rows, err := r.db.Query(membersQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		members = append(members, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	group.Members = members
	return group, nil
}

func (r *groupRepository) ListByUserID(userID uuid.UUID) ([]*domain.Group, error) {
	query := `
		SELECT DISTINCT g.id, g.name, g.created_by, g.created_at, g.updated_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY g.updated_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		group := &domain.Group{}
		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedBy,
			&group.CreatedAt,
			&group.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get members for this group
		membersQuery := `
			SELECT user_id
			FROM group_members
			WHERE group_id = $1
		`

		memberRows, err := r.db.Query(membersQuery, group.ID)
		if err != nil {
			return nil, err
		}

		var members []uuid.UUID
		for memberRows.Next() {
			var memberID uuid.UUID
			if err := memberRows.Scan(&memberID); err != nil {
				memberRows.Close()
				return nil, err
			}
			members = append(members, memberID)
		}
		memberRows.Close()

		if err = memberRows.Err(); err != nil {
			return nil, err
		}

		group.Members = members
		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *groupRepository) AddMember(groupID, userID uuid.UUID) error {
	query := `
		INSERT INTO group_members (group_id, user_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(query, groupID, userID)
	return err
}

func (r *groupRepository) RemoveMember(groupID, userID uuid.UUID) error {
	query := `
		DELETE FROM group_members
		WHERE group_id = $1 AND user_id = $2
	`

	result, err := r.db.Exec(query, groupID, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrNotGroupMember
	}

	return nil
}

func (r *groupRepository) IsMember(groupID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM group_members
			WHERE group_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRow(query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
