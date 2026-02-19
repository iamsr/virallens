package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// MessageRepositoryImpl implements domain.MessageRepository
type MessageRepositoryImpl struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
var _ domain.MessageRepository = (*MessageRepositoryImpl)(nil)

func NewMessageRepository(db *sql.DB) domain.MessageRepository {
	return &MessageRepositoryImpl{db: db}
}

func (r *MessageRepositoryImpl) Create(message *domain.Message) error {
	query := `
		INSERT INTO messages (id, sender_id, conversation_id, group_id, content, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Exec(
		query,
		message.ID,
		message.SenderID,
		message.ConversationID,
		message.GroupID,
		message.Content,
		message.Type,
		message.CreatedAt,
	)

	return err
}

func (r *MessageRepositoryImpl) GetByID(id uuid.UUID) (*domain.Message, error) {
	query := `
		SELECT id, sender_id, conversation_id, group_id, content, type, created_at
		FROM messages
		WHERE id = $1
	`

	message := &domain.Message{}
	err := r.db.QueryRow(query, id).Scan(
		&message.ID,
		&message.SenderID,
		&message.ConversationID,
		&message.GroupID,
		&message.Content,
		&message.Type,
		&message.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMessageNotFound
		}
		return nil, err
	}

	return message, nil
}

func (r *MessageRepositoryImpl) ListByConversationID(conversationID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	var query string
	var args []interface{}

	if cursor != nil {
		query = `
			SELECT id, sender_id, conversation_id, group_id, content, type, created_at
			FROM messages
			WHERE conversation_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{conversationID, cursor, limit}
	} else {
		query = `
			SELECT id, sender_id, conversation_id, group_id, content, type, created_at
			FROM messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{conversationID, limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ConversationID,
			&message.GroupID,
			&message.Content,
			&message.Type,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (r *MessageRepositoryImpl) ListByGroupID(groupID uuid.UUID, cursor *time.Time, limit int) ([]*domain.Message, error) {
	var query string
	var args []interface{}

	if cursor != nil {
		query = `
			SELECT id, sender_id, conversation_id, group_id, content, type, created_at
			FROM messages
			WHERE group_id = $1 AND created_at < $2
			ORDER BY created_at DESC
			LIMIT $3
		`
		args = []interface{}{groupID, cursor, limit}
	} else {
		query = `
			SELECT id, sender_id, conversation_id, group_id, content, type, created_at
			FROM messages
			WHERE group_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []interface{}{groupID, limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		message := &domain.Message{}
		err := rows.Scan(
			&message.ID,
			&message.SenderID,
			&message.ConversationID,
			&message.GroupID,
			&message.Content,
			&message.Type,
			&message.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}
