package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"dndproject/backend/internal/model"
)

type InviteRepository struct {
	db *sql.DB
}

func NewInviteRepository(db *sql.DB) *InviteRepository {
	return &InviteRepository{db: db}
}

var inviteSortColumns = []string{"created_at", "expires_at"}

func InviteSortColumns() []string { return inviteSortColumns }

const inviteColumns = `code, role, created_by, used_by, used_at, expires_at, created_at`

func scanInvite(row interface{ Scan(...any) error }) (model.Invite, error) {
	var i model.Invite
	err := row.Scan(&i.Code, &i.Role, &i.CreatedBy, &i.UsedBy, &i.UsedAt, &i.ExpiresAt, &i.CreatedAt)
	return i, err
}

func (r *InviteRepository) Create(ctx context.Context, code string, role model.Role, createdBy int64, expiresAt time.Time) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO invites (code, role, created_by, expires_at) VALUES (?, ?, ?, ?)`,
		code, role, createdBy, expiresAt)
	return err
}

func (r *InviteRepository) List(ctx context.Context, q model.Query) ([]model.Invite, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invites`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM invites ORDER BY %s %s LIMIT ? OFFSET ?`, inviteColumns, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, q.PageSize, q.Offset())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Invite
	for rows.Next() {
		i, err := scanInvite(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, rows.Err()
}

func (r *InviteRepository) Delete(ctx context.Context, code string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM invites WHERE code = ?`, code)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *InviteRepository) Get(ctx context.Context, code string) (model.Invite, error) {
	return scanInvite(r.db.QueryRowContext(ctx, `SELECT `+inviteColumns+` FROM invites WHERE code = ?`, code))
}
