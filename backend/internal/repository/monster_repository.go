package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

// MonsterRepository covers the bestiary entry, its skills and saves, and its
// actions. Like the character sheet, one screen reads all three together.
type MonsterRepository struct {
	db *sql.DB
}

func NewMonsterRepository(db *sql.DB) *MonsterRepository {
	return &MonsterRepository{db: db}
}

var monsterSortColumns = []string{"id", "name", "challenge_rating", "created_at", "updated_at"}

func MonsterSortColumns() []string { return monsterSortColumns }

const monsterColumns = `id, name, size, type, alignment, armor_class, hp_average, hp_formula, speed,
	strength, dexterity, constitution, intelligence, wisdom, charisma,
	challenge_rating, senses, languages, traits,
	damage_resistances, damage_vulnerabilities, damage_immunities, condition_immunities,
	color, image_url, token_size, created_at, updated_at`

func scanMonster(row interface{ Scan(...any) error }) (model.Monster, error) {
	var m model.Monster
	err := row.Scan(
		&m.ID, &m.Name, &m.Size, &m.Type, &m.Alignment, &m.ArmorClass, &m.HPAverage, &m.HPFormula, &m.Speed,
		&m.Abilities.Strength, &m.Abilities.Dexterity, &m.Abilities.Constitution,
		&m.Abilities.Intelligence, &m.Abilities.Wisdom, &m.Abilities.Charisma,
		&m.ChallengeRating, &m.Senses, &m.Languages, &m.Traits,
		&m.DamageResistances, &m.DamageVulnerabilities, &m.DamageImmunities, &m.ConditionImmunities,
		&m.Color, &m.ImageURL, &m.TokenSize,
		&m.CreatedAt, &m.UpdatedAt,
	)
	m.Modifiers = m.Abilities.Modifiers()
	return m, err
}

func (r *MonsterRepository) List(ctx context.Context, q model.Query) ([]model.Monster, int64, error) {
	where, args := "", []any{}
	if q.Search != "" {
		where = "WHERE name LIKE ? OR type LIKE ?"
		args = append(args, "%"+q.Search+"%", "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM monsters "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM monsters %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		monsterColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Monster
	for rows.Next() {
		m, err := scanMonster(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, m)
	}
	return items, total, rows.Err()
}

func (r *MonsterRepository) Get(ctx context.Context, id int64) (model.Monster, error) {
	return scanMonster(r.db.QueryRowContext(ctx, `SELECT `+monsterColumns+` FROM monsters WHERE id = ?`, id))
}

func (r *MonsterRepository) Create(ctx context.Context, in model.MonsterInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO monsters (name, size, type, alignment, armor_class, hp_average, hp_formula, speed,
			strength, dexterity, constitution, intelligence, wisdom, charisma,
			challenge_rating, senses, languages, traits,
			damage_resistances, damage_vulnerabilities, damage_immunities, condition_immunities,
			color, image_url, token_size)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.Name, in.Size, in.Type, in.Alignment, in.ArmorClass, in.HPAverage, in.HPFormula, in.Speed,
		in.Abilities.Strength, in.Abilities.Dexterity, in.Abilities.Constitution,
		in.Abilities.Intelligence, in.Abilities.Wisdom, in.Abilities.Charisma,
		in.ChallengeRating, in.Senses, in.Languages, in.Traits,
		in.DamageResistances, in.DamageVulnerabilities, in.DamageImmunities, in.ConditionImmunities,
		in.Color, in.ImageURL, in.TokenSize)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MonsterRepository) Update(ctx context.Context, id int64, in model.MonsterInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE monsters SET name = ?, size = ?, type = ?, alignment = ?, armor_class = ?,
			hp_average = ?, hp_formula = ?, speed = ?,
			strength = ?, dexterity = ?, constitution = ?, intelligence = ?, wisdom = ?, charisma = ?,
			challenge_rating = ?, senses = ?, languages = ?, traits = ?,
			damage_resistances = ?, damage_vulnerabilities = ?, damage_immunities = ?, condition_immunities = ?,
			color = ?, image_url = ?, token_size = ?
		 WHERE id = ?`,
		in.Name, in.Size, in.Type, in.Alignment, in.ArmorClass, in.HPAverage, in.HPFormula, in.Speed,
		in.Abilities.Strength, in.Abilities.Dexterity, in.Abilities.Constitution,
		in.Abilities.Intelligence, in.Abilities.Wisdom, in.Abilities.Charisma,
		in.ChallengeRating, in.Senses, in.Languages, in.Traits,
		in.DamageResistances, in.DamageVulnerabilities, in.DamageImmunities, in.ConditionImmunities,
		in.Color, in.ImageURL, in.TokenSize, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *MonsterRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM monsters WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Skills and saves

func (r *MonsterRepository) Proficiencies(ctx context.Context, monsterID int64) ([]model.MonsterProficiency, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT kind, name, modifier FROM monster_proficiencies WHERE monster_id = ? ORDER BY kind, name`, monsterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.MonsterProficiency{}
	for rows.Next() {
		var p model.MonsterProficiency
		if err := rows.Scan(&p.Kind, &p.Name, &p.Modifier); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// SetProficiency upserts one line, or removes it when keep is false.
func (r *MonsterRepository) SetProficiency(ctx context.Context, monsterID int64, p model.MonsterProficiency, keep bool) error {
	if !keep {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM monster_proficiencies WHERE monster_id = ? AND kind = ? AND name = ?`,
			monsterID, p.Kind, p.Name)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO monster_proficiencies (monster_id, kind, name, modifier) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE modifier = VALUES(modifier)`,
		monsterID, p.Kind, p.Name, p.Modifier)
	return err
}

// Actions

const monsterActionColumns = `id, monster_id, name, attack_bonus, damage, damage_type, description`

func (r *MonsterRepository) Actions(ctx context.Context, monsterID int64) ([]model.MonsterAction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+monsterActionColumns+` FROM monster_actions WHERE monster_id = ? ORDER BY id`, monsterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.MonsterAction{}
	for rows.Next() {
		var a model.MonsterAction
		if err := rows.Scan(&a.ID, &a.MonsterID, &a.Name, &a.AttackBonus, &a.Damage, &a.DamageType, &a.Description); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, rows.Err()
}

func (r *MonsterRepository) Action(ctx context.Context, id int64) (model.MonsterAction, error) {
	var a model.MonsterAction
	err := r.db.QueryRowContext(ctx, `SELECT `+monsterActionColumns+` FROM monster_actions WHERE id = ?`, id).
		Scan(&a.ID, &a.MonsterID, &a.Name, &a.AttackBonus, &a.Damage, &a.DamageType, &a.Description)
	return a, err
}

func (r *MonsterRepository) CreateAction(ctx context.Context, monsterID int64, in model.MonsterActionInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO monster_actions (monster_id, name, attack_bonus, damage, damage_type, description)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		monsterID, in.Name, in.AttackBonus, in.Damage, in.DamageType, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *MonsterRepository) UpdateAction(ctx context.Context, id int64, in model.MonsterActionInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE monster_actions SET name = ?, attack_bonus = ?, damage = ?, damage_type = ?, description = ?
		 WHERE id = ?`,
		in.Name, in.AttackBonus, in.Damage, in.DamageType, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *MonsterRepository) DeleteAction(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM monster_actions WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}
