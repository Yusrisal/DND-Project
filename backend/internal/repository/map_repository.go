package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type MapRepository struct {
	db *sql.DB
}

func NewMapRepository(db *sql.DB) *MapRepository {
	return &MapRepository{db: db}
}

var mapSortColumns = []string{"id", "name", "created_at", "updated_at"}

func MapSortColumns() []string { return mapSortColumns }

const mapColumns = `id, campaign_id, name, image_url, grid_size, width, height, fog_enabled, fog_revealed, weather, created_at, updated_at`

func scanMap(row interface{ Scan(...any) error }) (model.Map, error) {
	var m model.Map
	err := row.Scan(&m.ID, &m.CampaignID, &m.Name, &m.ImageURL, &m.GridSize, &m.Width, &m.Height, &m.FogEnabled, &m.FogRevealed, &m.Weather, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

// List filters by campaign when campaignID is non-zero.
func (r *MapRepository) List(ctx context.Context, campaignID int64, q model.Query) ([]model.Map, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM maps "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM maps %s ORDER BY %s %s LIMIT ? OFFSET ?`, mapColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Map
	for rows.Next() {
		m, err := scanMap(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, rows.Err()
}

func (r *MapRepository) Get(ctx context.Context, id int64) (model.Map, error) {
	return scanMap(r.db.QueryRowContext(ctx, `SELECT `+mapColumns+` FROM maps WHERE id = ?`, id))
}

func (r *MapRepository) Create(ctx context.Context, in model.MapInput) (int64, error) {
	revealed := in.FogRevealed
	if revealed == "" {
		revealed = "[]"
	}
	weather := in.Weather
	if weather == "" {
		weather = "clear"
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO maps (campaign_id, name, image_url, grid_size, width, height, fog_enabled, fog_revealed, weather) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.Name, in.ImageURL, in.GridSize, in.Width, in.Height, in.FogEnabled, revealed, weather)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MapRepository) Update(ctx context.Context, id int64, in model.MapInput) error {
	weather := in.Weather
	if weather == "" {
		weather = "clear"
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE maps SET name = ?, image_url = ?, grid_size = ?, width = ?, height = ?, weather = ? WHERE id = ?`,
		in.Name, in.ImageURL, in.GridSize, in.Width, in.Height, weather, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *MapRepository) UpdateFog(ctx context.Context, id int64, enabled bool, revealed string) error {
	if revealed == "" {
		revealed = "[]"
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE maps SET fog_enabled = ?, fog_revealed = ? WHERE id = ?`,
		enabled, revealed, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *MapRepository) SetWeather(ctx context.Context, id int64, weather string) error {
	if weather == "" {
		weather = "clear"
	}
	res, err := r.db.ExecContext(ctx, `UPDATE maps SET weather = ? WHERE id = ?`, weather, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *MapRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM maps WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// AoE Templates

func (r *MapRepository) ListAoETemplates(ctx context.Context, mapID int64) ([]model.AoETemplate, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, map_id, name, shape, size_ft, color, x, y, created_at FROM aoe_templates WHERE map_id = ? ORDER BY id ASC`,
		mapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.AoETemplate{}
	for rows.Next() {
		var t model.AoETemplate
		if err := rows.Scan(&t.ID, &t.MapID, &t.Name, &t.Shape, &t.SizeFt, &t.Color, &t.X, &t.Y, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (r *MapRepository) CreateAoETemplate(ctx context.Context, in model.AoETemplateInput) (model.AoETemplate, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO aoe_templates (map_id, name, shape, size_ft, color, x, y) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.MapID, in.Name, in.Shape, in.SizeFt, in.Color, in.X, in.Y)
	if err != nil {
		return model.AoETemplate{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.AoETemplate{}, err
	}
	var t model.AoETemplate
	err = r.db.QueryRowContext(ctx,
		`SELECT id, map_id, name, shape, size_ft, color, x, y, created_at FROM aoe_templates WHERE id = ?`, id).
		Scan(&t.ID, &t.MapID, &t.Name, &t.Shape, &t.SizeFt, &t.Color, &t.X, &t.Y, &t.CreatedAt)
	return t, err
}

func (r *MapRepository) DeleteAoETemplate(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM aoe_templates WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Map Objects (Doors, Chests, Traps, Markers)

const objectColumns = `id, map_id, kind, name, x, y, state, data, created_at, updated_at`

func scanMapObject(row interface{ Scan(...any) error }) (model.MapObject, error) {
	var o model.MapObject
	var data []byte
	err := row.Scan(&o.ID, &o.MapID, &o.Kind, &o.Name, &o.X, &o.Y, &o.State, &data, &o.CreatedAt, &o.UpdatedAt)
	o.Data = data
	return o, err
}

func (r *MapRepository) ListObjects(ctx context.Context, mapID int64) ([]model.MapObject, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+objectColumns+` FROM map_objects WHERE map_id = ? ORDER BY id ASC`, mapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.MapObject{}
	for rows.Next() {
		o, err := scanMapObject(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, o)
	}
	return list, rows.Err()
}

func (r *MapRepository) GetObject(ctx context.Context, id int64) (model.MapObject, error) {
	return scanMapObject(r.db.QueryRowContext(ctx, `SELECT `+objectColumns+` FROM map_objects WHERE id = ?`, id))
}

func (r *MapRepository) CreateObject(ctx context.Context, in model.MapObjectInput) (model.MapObject, error) {
	data := in.Data
	if len(data) == 0 {
		data = []byte(`{}`)
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO map_objects (map_id, kind, name, x, y, state, data) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		in.MapID, in.Kind, in.Name, in.X, in.Y, in.State, data)
	if err != nil {
		return model.MapObject{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return model.MapObject{}, err
	}
	return r.GetObject(ctx, id)
}

func (r *MapRepository) UpdateObject(ctx context.Context, id int64, in model.MapObjectInput) (model.MapObject, error) {
	data := in.Data
	if len(data) == 0 {
		data = []byte(`{}`)
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE map_objects SET name = ?, x = ?, y = ?, state = ?, data = ? WHERE id = ?`,
		in.Name, in.X, in.Y, in.State, data, id)
	if err != nil {
		return model.MapObject{}, err
	}
	if err := affected(res); err != nil {
		return model.MapObject{}, err
	}
	return r.GetObject(ctx, id)
}

func (r *MapRepository) DeleteObject(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM map_objects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
