package repository

import (
	"context"
	"database/sql"
	"errors"

	"dndproject/backend/internal/model"
)

type SkillRepository struct {
	db *sql.DB
}

func NewSkillRepository(db *sql.DB) *SkillRepository {
	return &SkillRepository{db: db}
}

const classSkillColumns = `id, class_name, name, tier, req_level, max_rank, parent_skill_id,
	damage, damage_type, cooldown_rounds, cost_resource, description, created_at, updated_at`

func scanClassSkill(row interface{ Scan(...any) error }) (model.ClassSkill, error) {
	var s model.ClassSkill
	err := row.Scan(&s.ID, &s.ClassName, &s.Name, &s.Tier, &s.ReqLevel, &s.MaxRank, &s.ParentSkillID,
		&s.Damage, &s.DamageType, &s.CooldownRounds, &s.CostResource, &s.Description,
		&s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *SkillRepository) ListClassSkills(ctx context.Context, className string) ([]model.ClassSkill, error) {
	query := `SELECT ` + classSkillColumns + ` FROM class_skills`
	args := []any{}
	if className != "" {
		query += ` WHERE class_name = ?`
		args = append(args, className)
	}
	query += ` ORDER BY tier, req_level, name`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.ClassSkill
	for rows.Next() {
		s, err := scanClassSkill(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, rows.Err()
}

func (r *SkillRepository) GetClassSkill(ctx context.Context, id int64) (model.ClassSkill, error) {
	return scanClassSkill(r.db.QueryRowContext(ctx,
		`SELECT `+classSkillColumns+` FROM class_skills WHERE id = ?`, id))
}

func (r *SkillRepository) CharacterSkills(ctx context.Context, characterID int64) ([]model.CharacterLearnedSkill, error) {
	query := `SELECT cs.id, cs.character_id, cs.skill_id, cs.current_rank, cs.is_active, cs.learned_at,
		s.id, s.class_name, s.name, s.tier, s.req_level, s.max_rank, s.parent_skill_id,
		s.damage, s.damage_type, s.cooldown_rounds, s.cost_resource, s.description,
		s.created_at, s.updated_at
		FROM character_skills cs
		JOIN class_skills s ON s.id = cs.skill_id
		WHERE cs.character_id = ?
		ORDER BY cs.is_active DESC, s.tier, s.name`

	rows, err := r.db.QueryContext(ctx, query, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.CharacterLearnedSkill
	for rows.Next() {
		var item model.CharacterLearnedSkill
		var s model.ClassSkill
		if err := rows.Scan(
			&item.ID, &item.CharacterID, &item.SkillID, &item.CurrentRank, &item.IsActive, &item.LearnedAt,
			&s.ID, &s.ClassName, &s.Name, &s.Tier, &s.ReqLevel, &s.MaxRank, &s.ParentSkillID,
			&s.Damage, &s.DamageType, &s.CooldownRounds, &s.CostResource, &s.Description,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Skill = s
		list = append(list, item)
	}
	return list, rows.Err()
}

func (r *SkillRepository) LearnSkill(ctx context.Context, characterID, skillID int64) error {
	// If already learned, upgrade rank
	var currentRank int
	err := r.db.QueryRowContext(ctx,
		`SELECT current_rank FROM character_skills WHERE character_id = ? AND skill_id = ?`,
		characterID, skillID).Scan(&currentRank)
	if err == nil {
		// Increment rank
		_, err = r.db.ExecContext(ctx,
			`UPDATE character_skills SET current_rank = current_rank + 1 WHERE character_id = ? AND skill_id = ?`,
			characterID, skillID)
		return err
	}

	// Count active skills to determine if this new skill can be active by default (max 10 active)
	var activeCount int
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM character_skills WHERE character_id = ? AND is_active = 1`, characterID).Scan(&activeCount)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	isActive := activeCount < 10
	_, err = r.db.ExecContext(ctx,
		`INSERT INTO character_skills (character_id, skill_id, current_rank, is_active)
		 VALUES (?, ?, 1, ?)`, characterID, skillID, isActive)
	return err
}

func (r *SkillRepository) SetSkillActive(ctx context.Context, characterID, skillID int64, active bool) error {
	if active {
		var activeCount int
		err := r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM character_skills WHERE character_id = ? AND is_active = 1 AND skill_id != ?`,
			characterID, skillID).Scan(&activeCount)
		if err != nil {
			return err
		}
		if activeCount >= 10 {
			return errors.New("cannot activate more than 10 skills in combat loadout")
		}
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE character_skills SET is_active = ? WHERE character_id = ? AND skill_id = ?`,
		active, characterID, skillID)
	if err != nil {
		return err
	}
	return affected(res)
}

// Loot Drops Queue
const pendingLootColumns = `id, campaign_id, map_id, monster_name, item_name, item_id,
	quantity, cp, sp, gp, status, reject_reason, assigned_to_character_id, created_at, updated_at`

func scanPendingLoot(row interface{ Scan(...any) error }) (model.PendingLootDrop, error) {
	var l model.PendingLootDrop
	err := row.Scan(&l.ID, &l.CampaignID, &l.MapID, &l.MonsterName, &l.ItemName, &l.ItemID,
		&l.Quantity, &l.CP, &l.SP, &l.GP, &l.Status, &l.RejectReason, &l.AssignedToCharacterID,
		&l.CreatedAt, &l.UpdatedAt)
	return l, err
}

func (r *SkillRepository) ListPendingLoot(ctx context.Context, campaignID int64) ([]model.PendingLootDrop, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+pendingLootColumns+` FROM pending_loot_drops WHERE campaign_id = ? AND status = 'pending' ORDER BY id DESC`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []model.PendingLootDrop
	for rows.Next() {
		l, err := scanPendingLoot(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, l)
	}
	return list, rows.Err()
}

func (r *SkillRepository) CreatePendingLoot(ctx context.Context, item model.PendingLootDrop) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO pending_loot_drops (campaign_id, map_id, monster_name, item_name, item_id,
			quantity, cp, sp, gp, status, assigned_to_character_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'pending', ?)`,
		item.CampaignID, item.MapID, item.MonsterName, item.ItemName, item.ItemID,
		item.Quantity, item.CP, item.SP, item.GP, item.AssignedToCharacterID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *SkillRepository) ResolvePendingLoot(ctx context.Context, id int64, approve bool, reason string) error {
	status := "approved"
	if !approve {
		status = "rejected"
	}
	res, err := r.db.ExecContext(ctx,
		`UPDATE pending_loot_drops SET status = ?, reject_reason = ? WHERE id = ?`,
		status, reason, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *SkillRepository) GetPendingLoot(ctx context.Context, id int64) (model.PendingLootDrop, error) {
	return scanPendingLoot(r.db.QueryRowContext(ctx,
		`SELECT `+pendingLootColumns+` FROM pending_loot_drops WHERE id = ?`, id))
}
