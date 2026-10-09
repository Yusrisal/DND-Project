package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type HandoutRepository struct {
	db *sql.DB
}

func NewHandoutRepository(db *sql.DB) *HandoutRepository {
	return &HandoutRepository{db: db}
}

var handoutSortColumns = []string{"id", "title", "created_at", "updated_at"}

func HandoutSortColumns() []string { return handoutSortColumns }

const handoutColumns = `id, campaign_id, title, content, image_url, is_revealed, created_at, updated_at`

func scanHandout(row interface{ Scan(...any) error }) (model.Handout, error) {
	var h model.Handout
	err := row.Scan(&h.ID, &h.CampaignID, &h.Title, &h.Content, &h.ImageURL, &h.IsRevealed, &h.CreatedAt, &h.UpdatedAt)
	return h, err
}

func (r *HandoutRepository) List(ctx context.Context, campaignID int64, revealedOnly bool, q model.Query) ([]model.Handout, int64, error) {
	where, args := "WHERE campaign_id = ?", []any{campaignID}
	if revealedOnly {
		where += " AND is_revealed = 1"
	}
	if q.Search != "" {
		where += " AND (title LIKE ? OR content LIKE ?)"
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM handouts "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM handouts %s ORDER BY %s %s LIMIT ? OFFSET ?`, handoutColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Handout
	for rows.Next() {
		h, err := scanHandout(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, h)
	}
	return items, total, rows.Err()
}

func (r *HandoutRepository) Get(ctx context.Context, id int64) (model.Handout, error) {
	return scanHandout(r.db.QueryRowContext(ctx, `SELECT `+handoutColumns+` FROM handouts WHERE id = ?`, id))
}

func (r *HandoutRepository) Create(ctx context.Context, in model.HandoutInput) (model.Handout, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO handouts (campaign_id, title, content, image_url, is_revealed) VALUES (?, ?, ?, ?, ?)`,
		in.CampaignID, in.Title, in.Content, in.ImageURL, in.IsRevealed)
	if err != nil {
		return model.Handout{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Handout{}, err
	}
	return r.Get(ctx, id)
}

func (r *HandoutRepository) Update(ctx context.Context, id int64, in model.HandoutInput) (model.Handout, error) {
	res, err := r.db.ExecContext(ctx,
		`UPDATE handouts SET title = ?, content = ?, image_url = ?, is_revealed = ? WHERE id = ?`,
		in.Title, in.Content, in.ImageURL, in.IsRevealed, id)
	if err != nil {
		return model.Handout{}, err
	}
	if err := affected(res); err != nil {
		return model.Handout{}, err
	}
	return r.Get(ctx, id)
}

func (r *HandoutRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM handouts WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
