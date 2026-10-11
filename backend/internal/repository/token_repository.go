package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

var tokenSortColumns = []string{"id", "name", "created_at", "updated_at"}

func TokenSortColumns() []string { return tokenSortColumns }

// A token linked to a sheet does not own its vitals: the sheet does. Reading
// them through the join means the board, the initiative list and the health
// bars all agree with the character sheet without a second round trip, and
// without the two copies drifting apart. An unlinked token keeps its own
// columns, which is what COALESCE falls back to.
const tokenColumns = `t.id, t.map_id, t.owner_user_id, t.character_id, t.monster_id, t.name, t.image_url,
	t.color, t.size,
	COALESCE(c.hp_current, t.hp_current), COALESCE(c.hp_max, t.hp_max),
	COALESCE(c.armor_class, t.armor_class),
	t.x, t.y, t.created_at, t.updated_at`

const tokenFrom = ` FROM tokens t LEFT JOIN characters c ON c.id = t.character_id `

func scanToken(row interface{ Scan(...any) error }) (model.Token, error) {
	var t model.Token
	err := row.Scan(&t.ID, &t.MapID, &t.OwnerUserID, &t.CharacterID, &t.MonsterID, &t.Name, &t.ImageURL, &t.Color, &t.Size,
		&t.HPCurrent, &t.HPMax, &t.ArmorClass, &t.X, &t.Y, &t.CreatedAt, &t.UpdatedAt)
	t.Conditions = []string{}
	return t, err
}

// ListByMap returns every token on a map. Tokens are bounded per map, so the
// board does not paginate; the paginated list is List.
func (r *TokenRepository) ListByMap(ctx context.Context, mapID int64) ([]model.Token, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT `+tokenColumns+tokenFrom+`WHERE t.map_id = ? ORDER BY t.id`, mapID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.Token{}
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

func (r *TokenRepository) List(ctx context.Context, mapID int64, q model.Query) ([]model.Token, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if mapID > 0 {
		where += " AND t.map_id = ?"
		args = append(args, mapID)
	}
	if q.Search != "" {
		where += " AND t.name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM tokens t "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// The sort column is qualified because the join makes a bare `name` or
	// `created_at` ambiguous. It is still allowlisted before it gets here.
	query := fmt.Sprintf(`SELECT %s%sORDER BY t.%s %s LIMIT ? OFFSET ?`,
		tokenColumns, tokenFrom+where+" ", q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Token
	for rows.Next() {
		t, err := scanToken(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, t)
	}
	return items, total, rows.Err()
}

func (r *TokenRepository) Get(ctx context.Context, id int64) (model.Token, error) {
	return scanToken(r.db.QueryRowContext(ctx, `SELECT `+tokenColumns+tokenFrom+`WHERE t.id = ?`, id))
}

func (r *TokenRepository) Create(ctx context.Context, in model.TokenInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO tokens (map_id, owner_user_id, character_id, monster_id, name, image_url, color, size,
			hp_current, hp_max, armor_class, x, y)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.MapID, in.OwnerUserID, in.CharacterID, in.MonsterID, in.Name, in.ImageURL, in.Color, in.Size,
		in.HPCurrent, in.HPMax, in.ArmorClass, in.X, in.Y)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *TokenRepository) Update(ctx context.Context, id int64, in model.TokenInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE tokens SET owner_user_id = ?, character_id = ?, monster_id = ?, name = ?, image_url = ?,
			color = ?, size = ?, hp_current = ?, hp_max = ?, armor_class = ?, x = ?, y = ? WHERE id = ?`,
		in.OwnerUserID, in.CharacterID, in.MonsterID, in.Name, in.ImageURL, in.Color, in.Size,
		in.HPCurrent, in.HPMax, in.ArmorClass, in.X, in.Y, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Move only writes the coordinates, so a drag does not race with an edit of the rest of the row.
func (r *TokenRepository) Move(ctx context.Context, id int64, x, y float64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE tokens SET x = ?, y = ? WHERE id = ?`, x, y, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *TokenRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM tokens WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Conditions

// LoadConditions fills in the status list for a set of tokens in one query,
// rather than one per token.
func (r *TokenRepository) LoadConditions(ctx context.Context, tokens []model.Token) error {
	if len(tokens) == 0 {
		return nil
	}

	ids := make([]any, len(tokens))
	placeholders := make([]byte, 0, len(tokens)*2)
	for i, token := range tokens {
		ids[i] = token.ID
		if i > 0 {
			placeholders = append(placeholders, ',')
		}
		placeholders = append(placeholders, '?')
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT token_id, condition_name FROM token_conditions WHERE token_id IN (`+string(placeholders)+`)`, ids...)
	if err != nil {
		return err
	}
	defer rows.Close()

	byToken := map[int64][]string{}
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		byToken[id] = append(byToken[id], name)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range tokens {
		if list, ok := byToken[tokens[i].ID]; ok {
			tokens[i].Conditions = list
		}
	}
	return nil
}

func (r *TokenRepository) SetCondition(ctx context.Context, tokenID int64, condition string, active bool) error {
	if !active {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM token_conditions WHERE token_id = ? AND condition_name = ?`, tokenID, condition)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT IGNORE INTO token_conditions (token_id, condition_name) VALUES (?, ?)`, tokenID, condition)
	return err
}

func (r *TokenRepository) SetConditionWithDuration(ctx context.Context, tokenID int64, condition string, active bool, duration int, notes string) error {
	if !active {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM token_conditions WHERE token_id = ? AND condition_name = ?`, tokenID, condition)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO token_conditions (token_id, condition_name, duration_rounds, remaining_rounds, notes)
		 VALUES (?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE duration_rounds = VALUES(duration_rounds), remaining_rounds = VALUES(remaining_rounds), notes = VALUES(notes)`,
		tokenID, condition, duration, duration, notes)
	return err
}

func (r *TokenRepository) DecrementConditions(ctx context.Context, mapID int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE token_conditions tc
		 JOIN tokens t ON tc.token_id = t.id
		 SET tc.remaining_rounds = tc.remaining_rounds - 1
		 WHERE t.map_id = ? AND tc.remaining_rounds > 0`, mapID)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx,
		`DELETE tc FROM token_conditions tc
		 JOIN tokens t ON tc.token_id = t.id
		 WHERE t.map_id = ? AND tc.duration_rounds > 0 AND tc.remaining_rounds <= 0`, mapID)
	return err
}

// Damage applies a delta to a token own hit points, clamped to its maximum.
// It is used only for tokens with no linked character sheet.
func (r *TokenRepository) Damage(ctx context.Context, id int64, delta int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE tokens SET hp_current = GREATEST(0, LEAST(hp_max, hp_current + ?)) WHERE id = ?`, delta, id)
	if err != nil {
		return err
	}
	return affected(res)
}
