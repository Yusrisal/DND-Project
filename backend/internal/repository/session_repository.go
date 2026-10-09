package repository

import (
	"context"
	"database/sql"
	"time"

	"dndproject/backend/internal/model"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, tokenHash string, userID int64, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`, tokenHash, userID, expiresAt)
	return err
}

// User resolves a session cookie to its owner in one query, and treats an
// expired row as no row at all.
func (r *SessionRepository) User(ctx context.Context, tokenHash string) (model.User, error) {
	return scanUser(r.db.QueryRowContext(ctx,
		`SELECT u.id, u.email, u.display_name, u.role, u.created_at, u.updated_at
		   FROM sessions s
		   JOIN users u ON u.id = s.user_id
		  WHERE s.token_hash = ? AND s.expires_at > NOW()`, tokenHash))
}

func (r *SessionRepository) Delete(ctx context.Context, tokenHash string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE token_hash = ?`, tokenHash)
	return err
}

// DeleteExpired keeps the table from growing without bound.
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM sessions WHERE expires_at <= NOW()`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}
