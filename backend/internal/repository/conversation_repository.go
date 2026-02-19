package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

type conversationRepository struct {
	db *sql.DB
}

// NewConversationRepository creates a new conversation repository
func NewConversationRepository(db *sql.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(conv *domain.Conversation) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Insert conversation
	query := `
		INSERT INTO conversations (id, created_at, updated_at)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(query, conv.ID, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return err
	}

	// Insert participants
	for _, participantID := range conv.Participants {
		participantQuery := `
			INSERT INTO conversation_participants (conversation_id, user_id)
			VALUES ($1, $2)
		`
		_, err = tx.Exec(participantQuery, conv.ID, participantID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *conversationRepository) GetByID(id uuid.UUID) (*domain.Conversation, error) {
	// Get conversation
	query := `
		SELECT id, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	conv := &domain.Conversation{}
	err := r.db.QueryRow(query, id).Scan(
		&conv.ID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConversationNotFound
		}
		return nil, err
	}

	// Get participants
	participantsQuery := `
		SELECT user_id
		FROM conversation_participants
		WHERE conversation_id = $1
	`

	rows, err := r.db.Query(participantsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []uuid.UUID
	for rows.Next() {
		var userID uuid.UUID
		if err := rows.Scan(&userID); err != nil {
			return nil, err
		}
		participants = append(participants, userID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	conv.Participants = participants
	return conv, nil
}

func (r *conversationRepository) GetByParticipants(user1ID, user2ID uuid.UUID) (*domain.Conversation, error) {
	query := `
		SELECT c.id, c.created_at, c.updated_at
		FROM conversations c
		WHERE c.id IN (
			SELECT cp1.conversation_id
			FROM conversation_participants cp1
			WHERE cp1.user_id = $1
			INTERSECT
			SELECT cp2.conversation_id
			FROM conversation_participants cp2
			WHERE cp2.user_id = $2
		)
		AND (
			SELECT COUNT(*)
			FROM conversation_participants
			WHERE conversation_id = c.id
		) = 2
		LIMIT 1
	`

	conv := &domain.Conversation{}
	err := r.db.QueryRow(query, user1ID, user2ID).Scan(
		&conv.ID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConversationNotFound
		}
		return nil, err
	}

	// Get participants
	conv.Participants = []uuid.UUID{user1ID, user2ID}

	return conv, nil
}

func (r *conversationRepository) ListByUserID(userID uuid.UUID) ([]*domain.Conversation, error) {
	query := `
		SELECT DISTINCT c.id, c.created_at, c.updated_at
		FROM conversations c
		JOIN conversation_participants cp ON c.id = cp.conversation_id
		WHERE cp.user_id = $1
		ORDER BY c.updated_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*domain.Conversation
	for rows.Next() {
		conv := &domain.Conversation{}
		err := rows.Scan(
			&conv.ID,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get participants for this conversation
		participantsQuery := `
			SELECT user_id
			FROM conversation_participants
			WHERE conversation_id = $1
		`

		participantRows, err := r.db.Query(participantsQuery, conv.ID)
		if err != nil {
			return nil, err
		}

		var participants []uuid.UUID
		for participantRows.Next() {
			var participantID uuid.UUID
			if err := participantRows.Scan(&participantID); err != nil {
				participantRows.Close()
				return nil, err
			}
			participants = append(participants, participantID)
		}
		participantRows.Close()

		if err = participantRows.Err(); err != nil {
			return nil, err
		}

		conv.Participants = participants
		conversations = append(conversations, conv)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return conversations, nil
}

func (r *conversationRepository) AddParticipant(conversationID, userID uuid.UUID) error {
	query := `
		INSERT INTO conversation_participants (conversation_id, user_id)
		VALUES ($1, $2)
	`

	_, err := r.db.Exec(query, conversationID, userID)
	return err
}
