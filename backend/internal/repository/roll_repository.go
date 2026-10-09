package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"slices"

	"dndproject/backend/internal/model"
)

type RollRepository struct {
	db *sql.DB
}

func NewRollRepository(db *sql.DB) *RollRepository {
	return &RollRepository{db: db}
}

const rollColumns = `id, campaign_id, user_id, character_id, author_name, label, expression, detail, total, created_at`

func scanRoll(row interface{ Scan(...any) error }) (model.DiceRoll, error) {
	var r model.DiceRoll
	var detail []byte
	err := row.Scan(&r.ID, &r.CampaignID, &r.UserID, &r.CharacterID, &r.AuthorName,
		&r.Label, &r.Expression, &detail, &r.Total, &r.CreatedAt)
	r.Detail = json.RawMessage(detail)
	return r, err
}

func (r *RollRepository) Create(ctx context.Context, roll model.DiceRoll) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO dice_rolls (campaign_id, user_id, character_id, author_name, label, expression, detail, total)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		roll.CampaignID, roll.UserID, roll.CharacterID, roll.AuthorName,
		roll.Label, roll.Expression, []byte(roll.Detail), roll.Total)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RollRepository) Get(ctx context.Context, id int64) (model.DiceRoll, error) {
	return scanRoll(r.db.QueryRowContext(ctx, `SELECT `+rollColumns+` FROM dice_rolls WHERE id = ?`, id))
}

// Recent returns the tail of the roll log in chronological order.
// If retentionDays > 0, only rolls within that window are returned.
func (r *RollRepository) Recent(ctx context.Context, campaignID int64, limit int, retentionDays int) ([]model.DiceRoll, error) {
	where := "WHERE campaign_id = ?"
	args := []any{campaignID}
	if retentionDays > 0 {
		where += " AND created_at >= NOW() - INTERVAL ? DAY"
		args = append(args, retentionDays)
	}

	query := `SELECT ` + rollColumns + ` FROM dice_rolls ` + where + ` ORDER BY id DESC LIMIT ?`
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.DiceRoll{}
	for rows.Next() {
		roll, err := scanRoll(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, roll)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	slices.Reverse(items)
	return items, nil
}
