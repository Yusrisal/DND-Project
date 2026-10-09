package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type CampaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

var campaignSortColumns = []string{"id", "name", "created_at", "updated_at"}

func CampaignSortColumns() []string { return campaignSortColumns }

const campaignColumns = `id, name, description, is_paused, pause_message, chat_retention_days, created_at, updated_at`

func scanCampaign(row interface{ Scan(...any) error }) (model.Campaign, error) {
	var c model.Campaign
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.IsPaused, &c.PauseMessage, &c.ChatRetentionDays, &c.CreatedAt, &c.UpdatedAt)
	return c, err
}

func (r *CampaignRepository) List(ctx context.Context, q model.Query) ([]model.Campaign, int64, error) {
	where, args := "", []any{}
	if q.Search != "" {
		where = "WHERE name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaigns "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(
		`SELECT %s FROM campaigns %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		campaignColumns, where, q.SortBy, q.SortDir,
	)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Campaign
	for rows.Next() {
		c, err := scanCampaign(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}

func (r *CampaignRepository) Get(ctx context.Context, id int64) (model.Campaign, error) {
	return scanCampaign(r.db.QueryRowContext(ctx, `SELECT `+campaignColumns+` FROM campaigns WHERE id = ?`, id))
}

func (r *CampaignRepository) Create(ctx context.Context, in model.CampaignInput) (int64, error) {
	msg := in.PauseMessage
	if msg == "" {
		msg = "Session Paused by DM"
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO campaigns (name, description, is_paused, pause_message, chat_retention_days) VALUES (?, ?, ?, ?, ?)`,
		in.Name, in.Description, in.IsPaused, msg, in.ChatRetentionDays)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CampaignRepository) Update(ctx context.Context, id int64, in model.CampaignInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET name = ?, description = ?, chat_retention_days = ? WHERE id = ?`,
		in.Name, in.Description, in.ChatRetentionDays, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CampaignRepository) SetPause(ctx context.Context, id int64, isPaused bool, message string) error {
	if message == "" {
		message = "Session Paused by DM"
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET is_paused = ?, pause_message = ? WHERE id = ?`,
		isPaused, message, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CampaignRepository) SetRetention(ctx context.Context, id int64, days int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE campaigns SET chat_retention_days = ? WHERE id = ?`,
		days, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CampaignRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM campaigns WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
