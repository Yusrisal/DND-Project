package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type SavepointRepository struct {
	db *sql.DB
}

func NewSavepointRepository(db *sql.DB) *SavepointRepository {
	return &SavepointRepository{db: db}
}

var savepointSortColumns = []string{"id", "name", "created_at"}

func SavepointSortColumns() []string { return savepointSortColumns }

const savepointColumns = `id, campaign_id, map_id, name, notes, snapshot_data, created_at`

func scanSavepoint(row interface{ Scan(...any) error }) (model.Savepoint, error) {
	var s model.Savepoint
	err := row.Scan(&s.ID, &s.CampaignID, &s.MapID, &s.Name, &s.Notes, &s.SnapshotData, &s.CreatedAt)
	return s, err
}

func (r *SavepointRepository) List(ctx context.Context, campaignID int64, q model.Query) ([]model.Savepoint, int64, error) {
	where, args := "WHERE campaign_id = ?", []any{campaignID}
	if q.Search != "" {
		where += " AND (name LIKE ? OR notes LIKE ?)"
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaign_savepoints "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM campaign_savepoints %s ORDER BY %s %s LIMIT ? OFFSET ?`, savepointColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Savepoint
	for rows.Next() {
		s, err := scanSavepoint(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, s)
	}
	return items, total, rows.Err()
}

func (r *SavepointRepository) Get(ctx context.Context, id int64) (model.Savepoint, error) {
	return scanSavepoint(r.db.QueryRowContext(ctx, `SELECT `+savepointColumns+` FROM campaign_savepoints WHERE id = ?`, id))
}

func (r *SavepointRepository) Create(ctx context.Context, campaignID int64, mapID *int64, name, notes, snapshotData string) (model.Savepoint, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO campaign_savepoints (campaign_id, map_id, name, notes, snapshot_data) VALUES (?, ?, ?, ?, ?)`,
		campaignID, mapID, name, notes, snapshotData)
	if err != nil {
		return model.Savepoint{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.Savepoint{}, err
	}
	return r.Get(ctx, id)
}

func (r *SavepointRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM campaign_savepoints WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
