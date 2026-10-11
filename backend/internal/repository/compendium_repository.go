package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

type CompendiumRepository struct {
	db *sql.DB
}

func NewCompendiumRepository(db *sql.DB) *CompendiumRepository {
	return &CompendiumRepository{db: db}
}

var compendiumSortColumns = []string{"id", "name", "created_at", "updated_at"}

func CompendiumSortColumns() []string { return compendiumSortColumns }

// Environments

const environmentColumns = `id, campaign_id, map_id, name, kind, difficulty_dc, damage, damage_type, effect_description, active, created_at, updated_at`

func scanEnvironment(row interface{ Scan(...any) error }) (model.Environment, error) {
	var e model.Environment
	err := row.Scan(&e.ID, &e.CampaignID, &e.MapID, &e.Name, &e.Kind, &e.DifficultyDC, &e.Damage, &e.DamageType,
		&e.EffectDescription, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	return e, err
}

func (r *CompendiumRepository) ListEnvironments(ctx context.Context, campaignID, mapID int64, q model.Query) ([]model.Environment, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if mapID > 0 {
		where += " AND (map_id = ? OR map_id IS NULL)"
		args = append(args, mapID)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM environments "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM environments %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		environmentColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Environment
	for rows.Next() {
		e, err := scanEnvironment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, e)
	}
	return items, total, rows.Err()
}

func (r *CompendiumRepository) GetEnvironment(ctx context.Context, id int64) (model.Environment, error) {
	return scanEnvironment(r.db.QueryRowContext(ctx, `SELECT `+environmentColumns+` FROM environments WHERE id = ?`, id))
}

func (r *CompendiumRepository) CreateEnvironment(ctx context.Context, in model.EnvironmentInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO environments (campaign_id, map_id, name, kind, difficulty_dc, damage, damage_type, effect_description, active)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.MapID, in.Name, in.Kind, in.DifficultyDC, in.Damage, in.DamageType, in.EffectDescription, in.Active)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CompendiumRepository) UpdateEnvironment(ctx context.Context, id int64, in model.EnvironmentInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE environments SET map_id = ?, name = ?, kind = ?, difficulty_dc = ?, damage = ?, damage_type = ?, effect_description = ?, active = ?
		 WHERE id = ?`,
		in.MapID, in.Name, in.Kind, in.DifficultyDC, in.Damage, in.DamageType, in.EffectDescription, in.Active, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CompendiumRepository) DeleteEnvironment(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM environments WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Compendium Items

const compendiumItemColumns = `id, campaign_id, name, category, rarity, cost_cp, weight, attack_bonus, damage, damage_type, description,
	req_level, req_class, req_str, req_dex, req_con, req_int, req_wis, req_cha, teaches_skill_id,
	set_name, set_piece, set_bonus_2pc, set_bonus_4pc,
	created_at, updated_at`

func scanCompendiumItem(row interface{ Scan(...any) error }) (model.CompendiumItem, error) {
	var i model.CompendiumItem
	err := row.Scan(&i.ID, &i.CampaignID, &i.Name, &i.Category, &i.Rarity, &i.CostCP, &i.Weight,
		&i.AttackBonus, &i.Damage, &i.DamageType, &i.Description,
		&i.ReqLevel, &i.ReqClass, &i.ReqStr, &i.ReqDex, &i.ReqCon, &i.ReqInt, &i.ReqWis, &i.ReqCha, &i.TeachesSkillID,
		&i.SetName, &i.SetPiece, &i.SetBonus2pc, &i.SetBonus4pc,
		&i.CreatedAt, &i.UpdatedAt)
	return i, err
}

func (r *CompendiumRepository) ListItems(ctx context.Context, campaignID int64, category string, q model.Query) ([]model.CompendiumItem, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND (campaign_id = ? OR campaign_id IS NULL)"
		args = append(args, campaignID)
	}
	if category != "" {
		where += " AND category = ?"
		args = append(args, category)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compendium_items "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM compendium_items %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		compendiumItemColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.CompendiumItem
	for rows.Next() {
		i, err := scanCompendiumItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, i)
	}
	return items, total, rows.Err()
}

func (r *CompendiumRepository) GetItem(ctx context.Context, id int64) (model.CompendiumItem, error) {
	return scanCompendiumItem(r.db.QueryRowContext(ctx, `SELECT `+compendiumItemColumns+` FROM compendium_items WHERE id = ?`, id))
}

func (r *CompendiumRepository) CreateItem(ctx context.Context, in model.CompendiumItemInput) (int64, error) {
	if in.ReqLevel <= 0 {
		in.ReqLevel = 1
	}
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO compendium_items (campaign_id, name, category, rarity, cost_cp, weight, attack_bonus, damage, damage_type, description,
			req_level, req_class, req_str, req_dex, req_con, req_int, req_wis, req_cha, teaches_skill_id,
			set_name, set_piece, set_bonus_2pc, set_bonus_4pc)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.Name, in.Category, in.Rarity, in.CostCP, in.Weight, in.AttackBonus, in.Damage, in.DamageType, in.Description,
		in.ReqLevel, in.ReqClass, in.ReqStr, in.ReqDex, in.ReqCon, in.ReqInt, in.ReqWis, in.ReqCha, in.TeachesSkillID,
		in.SetName, in.SetPiece, in.SetBonus2pc, in.SetBonus4pc)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CompendiumRepository) UpdateItem(ctx context.Context, id int64, in model.CompendiumItemInput) error {
	if in.ReqLevel <= 0 {
		in.ReqLevel = 1
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE compendium_items SET name = ?, category = ?, rarity = ?, cost_cp = ?, weight = ?, attack_bonus = ?, damage = ?, damage_type = ?, description = ?,
			req_level = ?, req_class = ?, req_str = ?, req_dex = ?, req_con = ?, req_int = ?, req_wis = ?, req_cha = ?, teaches_skill_id = ?,
			set_name = ?, set_piece = ?, set_bonus_2pc = ?, set_bonus_4pc = ?
		 WHERE id = ?`,
		in.Name, in.Category, in.Rarity, in.CostCP, in.Weight, in.AttackBonus, in.Damage, in.DamageType, in.Description,
		in.ReqLevel, in.ReqClass, in.ReqStr, in.ReqDex, in.ReqCon, in.ReqInt, in.ReqWis, in.ReqCha, in.TeachesSkillID,
		in.SetName, in.SetPiece, in.SetBonus2pc, in.SetBonus4pc, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CompendiumRepository) DeleteItem(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM compendium_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Monster Drops

const monsterDropColumns = `id, monster_id, item_name, item_id, drop_rate, min_quantity, max_quantity, cp, sp, gp, created_at, updated_at`

func scanMonsterDrop(row interface{ Scan(...any) error }) (model.MonsterDrop, error) {
	var d model.MonsterDrop
	err := row.Scan(&d.ID, &d.MonsterID, &d.ItemName, &d.ItemID, &d.DropRate, &d.MinQuantity, &d.MaxQuantity,
		&d.CP, &d.SP, &d.GP, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *CompendiumRepository) ListMonsterDrops(ctx context.Context, monsterID int64) ([]model.MonsterDrop, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+monsterDropColumns+` FROM monster_drops WHERE monster_id = ? ORDER BY drop_rate DESC, item_name`, monsterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	drops := []model.MonsterDrop{}
	for rows.Next() {
		d, err := scanMonsterDrop(rows)
		if err != nil {
			return nil, err
		}
		drops = append(drops, d)
	}
	return drops, rows.Err()
}

func (r *CompendiumRepository) CreateMonsterDrop(ctx context.Context, in model.MonsterDropInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO monster_drops (monster_id, item_name, item_id, drop_rate, min_quantity, max_quantity, cp, sp, gp)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.MonsterID, in.ItemName, in.ItemID, in.DropRate, in.MinQuantity, in.MaxQuantity, in.CP, in.SP, in.GP)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CompendiumRepository) DeleteMonsterDrop(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM monster_drops WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Compendium Spells

const compendiumSpellColumns = `id, campaign_id, name, level, school, casting_time, range_area, duration, damage, damage_type, save_ability, description, created_at, updated_at`

func scanCompendiumSpell(row interface{ Scan(...any) error }) (model.CompendiumSpell, error) {
	var s model.CompendiumSpell
	err := row.Scan(&s.ID, &s.CampaignID, &s.Name, &s.Level, &s.School, &s.CastingTime, &s.RangeArea,
		&s.Duration, &s.Damage, &s.DamageType, &s.SaveAbility, &s.Description, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *CompendiumRepository) ListSpells(ctx context.Context, campaignID int64, level int, q model.Query) ([]model.CompendiumSpell, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND (campaign_id = ? OR campaign_id IS NULL)"
		args = append(args, campaignID)
	}
	if level >= 0 {
		where += " AND level = ?"
		args = append(args, level)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM compendium_spells "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM compendium_spells %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		compendiumSpellColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var spells []model.CompendiumSpell
	for rows.Next() {
		s, err := scanCompendiumSpell(rows)
		if err != nil {
			return nil, 0, err
		}
		spells = append(spells, s)
	}
	return spells, total, rows.Err()
}

func (r *CompendiumRepository) GetSpell(ctx context.Context, id int64) (model.CompendiumSpell, error) {
	return scanCompendiumSpell(r.db.QueryRowContext(ctx, `SELECT `+compendiumSpellColumns+` FROM compendium_spells WHERE id = ?`, id))
}

func (r *CompendiumRepository) CreateSpell(ctx context.Context, in model.CompendiumSpellInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO compendium_spells (campaign_id, name, level, school, casting_time, range_area, duration, damage, damage_type, save_ability, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.Name, in.Level, in.School, in.CastingTime, in.RangeArea, in.Duration, in.Damage, in.DamageType, in.SaveAbility, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CompendiumRepository) UpdateSpell(ctx context.Context, id int64, in model.CompendiumSpellInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE compendium_spells SET name = ?, level = ?, school = ?, casting_time = ?, range_area = ?, duration = ?, damage = ?, damage_type = ?, save_ability = ?, description = ?
		 WHERE id = ?`,
		in.Name, in.Level, in.School, in.CastingTime, in.RangeArea, in.Duration, in.Damage, in.DamageType, in.SaveAbility, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CompendiumRepository) DeleteSpell(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM compendium_spells WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Status Effects

const statusEffectColumns = `id, campaign_id, name, kind, default_duration_rounds, color, effect_description, created_at, updated_at`

func scanStatusEffect(row interface{ Scan(...any) error }) (model.StatusEffect, error) {
	var s model.StatusEffect
	err := row.Scan(&s.ID, &s.CampaignID, &s.Name, &s.Kind, &s.DefaultDurationRounds, &s.Color, &s.EffectDescription, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *CompendiumRepository) ListStatusEffects(ctx context.Context, campaignID int64, kind string, q model.Query) ([]model.StatusEffect, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND (campaign_id = ? OR campaign_id IS NULL)"
		args = append(args, campaignID)
	}
	if kind != "" {
		where += " AND kind = ?"
		args = append(args, kind)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM status_effects "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM status_effects %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		statusEffectColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var effects []model.StatusEffect
	for rows.Next() {
		s, err := scanStatusEffect(rows)
		if err != nil {
			return nil, 0, err
		}
		effects = append(effects, s)
	}
	return effects, total, rows.Err()
}

func (r *CompendiumRepository) GetStatusEffect(ctx context.Context, id int64) (model.StatusEffect, error) {
	return scanStatusEffect(r.db.QueryRowContext(ctx, `SELECT `+statusEffectColumns+` FROM status_effects WHERE id = ?`, id))
}

func (r *CompendiumRepository) CreateStatusEffect(ctx context.Context, in model.StatusEffectInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO status_effects (campaign_id, name, kind, default_duration_rounds, color, effect_description)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.Name, in.Kind, in.DefaultDurationRounds, in.Color, in.EffectDescription)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CompendiumRepository) UpdateStatusEffect(ctx context.Context, id int64, in model.StatusEffectInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE status_effects SET name = ?, kind = ?, default_duration_rounds = ?, color = ?, effect_description = ?
		 WHERE id = ?`,
		in.Name, in.Kind, in.DefaultDurationRounds, in.Color, in.EffectDescription, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CompendiumRepository) DeleteStatusEffect(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM status_effects WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
