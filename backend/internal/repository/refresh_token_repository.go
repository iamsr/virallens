package repository

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/virallens/backend/internal/domain"
)

// RefreshTokenRepositoryImpl implements domain.RefreshTokenRepository
type RefreshTokenRepositoryImpl struct {
	db *sql.DB
}

// NewRefreshTokenRepository creates a new refresh token repository
var _ domain.RefreshTokenRepository = (*RefreshTokenRepositoryImpl)(nil)

func NewRefreshTokenRepository(db *sql.DB) domain.RefreshTokenRepository {
	return &RefreshTokenRepositoryImpl{db: db}
}

func (r *RefreshTokenRepositoryImpl) Create(token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := r.db.Exec(
		query,
		token.ID,
		token.UserID,
		token.Token,
		token.ExpiresAt,
		token.CreatedAt,
	)

	return err
}

func (r *RefreshTokenRepositoryImpl) GetByToken(token string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, token, expires_at, created_at
		FROM refresh_tokens
		WHERE token = $1
	`

	refreshToken := &domain.RefreshToken{}
	err := r.db.QueryRow(query, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.ExpiresAt,
		&refreshToken.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}

	return refreshToken, nil
}

func (r *RefreshTokenRepositoryImpl) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(query, userID)
	return err
}

func (r *RefreshTokenRepositoryImpl) DeleteExpired() error {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	_, err := r.db.Exec(query)
	return err
}
