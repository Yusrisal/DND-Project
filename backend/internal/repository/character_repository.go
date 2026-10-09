package repository

import (
	"context"
	"database/sql"
	"fmt"

	"dndproject/backend/internal/model"
)

// CharacterRepository covers the whole sheet: the character row, its items, its
// spells and its spell slots. One screen reads and writes them together, so
// splitting them across four repositories would buy nothing.
type CharacterRepository struct {
	db *sql.DB
}

func NewCharacterRepository(db *sql.DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

var characterSortColumns = []string{"id", "name", "level", "status", "created_at", "updated_at"}

func CharacterSortColumns() []string { return characterSortColumns }

const characterColumns = `id, campaign_id, user_id, status, name, class, race, level,
	hp_current, hp_max, hp_temp, armor_class, speed, proficiency_bonus,
	cp, sp, ep, gp, pp,
	strength, dexterity, constitution, intelligence, wisdom, charisma,
	notes, death_save_successes, death_save_failures, hit_dice_total, hit_dice_used, hit_dice_die,
	damage_resistances, damage_vulnerabilities, damage_immunities, condition_immunities,
	created_at, updated_at`

func scanCharacter(row interface{ Scan(...any) error }) (model.Character, error) {
	var c model.Character
	err := row.Scan(
		&c.ID, &c.CampaignID, &c.UserID, &c.Status, &c.Name, &c.Class, &c.Race, &c.Level,
		&c.HPCurrent, &c.HPMax, &c.HPTemp, &c.ArmorClass, &c.Speed, &c.ProficiencyBonus,
		&c.CP, &c.SP, &c.EP, &c.GP, &c.PP,
		&c.Abilities.Strength, &c.Abilities.Dexterity, &c.Abilities.Constitution,
		&c.Abilities.Intelligence, &c.Abilities.Wisdom, &c.Abilities.Charisma,
		&c.Notes, &c.DeathSaveSuccesses, &c.DeathSaveFailures, &c.HitDiceTotal, &c.HitDiceUsed, &c.HitDiceDie,
		&c.DamageResistances, &c.DamageVulnerabilities, &c.DamageImmunities, &c.ConditionImmunities,
		&c.CreatedAt, &c.UpdatedAt,
	)
	c.Modifiers = c.Abilities.Modifiers()
	return c, err
}

// List filters by campaign and, when userID is non-zero, to the sheets one
// player owns.
func (r *CharacterRepository) List(ctx context.Context, campaignID, userID int64, q model.Query) ([]model.Character, int64, error) {
	where, args := "WHERE 1 = 1", []any{}
	if campaignID > 0 {
		where += " AND campaign_id = ?"
		args = append(args, campaignID)
	}
	if userID > 0 {
		where += " AND user_id = ?"
		args = append(args, userID)
	}
	if q.Search != "" {
		where += " AND name LIKE ?"
		args = append(args, "%"+q.Search+"%")
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM characters "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`SELECT %s FROM characters %s ORDER BY %s %s LIMIT ? OFFSET ?`,
		characterColumns, where, q.SortBy, q.SortDir)
	rows, err := r.db.QueryContext(ctx, query, append(args, q.PageSize, q.Offset())...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.Character
	for rows.Next() {
		c, err := scanCharacter(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, c)
	}
	return items, total, rows.Err()
}

func (r *CharacterRepository) Get(ctx context.Context, id int64) (model.Character, error) {
	return scanCharacter(r.db.QueryRowContext(ctx, `SELECT `+characterColumns+` FROM characters WHERE id = ?`, id))
}

func (r *CharacterRepository) Create(ctx context.Context, in model.CharacterInput) (int64, error) {
	if in.Status == "" {
		in.Status = "active"
	}
	if in.HitDiceDie == "" {
		in.HitDiceDie = "d8"
	}
	if in.HitDiceTotal <= 0 {
		in.HitDiceTotal = max(1, in.Level)
	}

	res, err := r.db.ExecContext(ctx,
		`INSERT INTO characters (campaign_id, user_id, status, name, class, race, level,
			hp_current, hp_max, hp_temp, armor_class, speed, proficiency_bonus,
			cp, sp, ep, gp, pp,
			strength, dexterity, constitution, intelligence, wisdom, charisma,
			notes, death_save_successes, death_save_failures, hit_dice_total, hit_dice_used, hit_dice_die,
			damage_resistances, damage_vulnerabilities, damage_immunities, condition_immunities)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		in.CampaignID, in.UserID, in.Status, in.Name, in.Class, in.Race, in.Level,
		in.HPCurrent, in.HPMax, in.HPTemp, in.ArmorClass, in.Speed, in.ProficiencyBonus,
		in.CP, in.SP, in.EP, in.GP, in.PP,
		in.Abilities.Strength, in.Abilities.Dexterity, in.Abilities.Constitution,
		in.Abilities.Intelligence, in.Abilities.Wisdom, in.Abilities.Charisma,
		in.Notes, in.DeathSaveSuccesses, in.DeathSaveFailures, in.HitDiceTotal, in.HitDiceUsed, in.HitDiceDie,
		in.DamageResistances, in.DamageVulnerabilities, in.DamageImmunities, in.ConditionImmunities)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Update leaves campaign_id alone: a character does not move between campaigns.
func (r *CharacterRepository) Update(ctx context.Context, id int64, in model.CharacterInput) error {
	if in.Status == "" {
		in.Status = "active"
	}
	if in.HitDiceDie == "" {
		in.HitDiceDie = "d8"
	}
	if in.HitDiceTotal <= 0 {
		in.HitDiceTotal = max(1, in.Level)
	}

	res, err := r.db.ExecContext(ctx,
		`UPDATE characters SET user_id = ?, status = ?, name = ?, class = ?, race = ?, level = ?,
			hp_current = ?, hp_max = ?, hp_temp = ?, armor_class = ?, speed = ?, proficiency_bonus = ?,
			cp = ?, sp = ?, ep = ?, gp = ?, pp = ?,
			strength = ?, dexterity = ?, constitution = ?, intelligence = ?, wisdom = ?, charisma = ?,
			notes = ?, death_save_successes = ?, death_save_failures = ?,
			hit_dice_total = ?, hit_dice_used = ?, hit_dice_die = ?,
			damage_resistances = ?, damage_vulnerabilities = ?, damage_immunities = ?, condition_immunities = ?
		 WHERE id = ?`,
		in.UserID, in.Status, in.Name, in.Class, in.Race, in.Level,
		in.HPCurrent, in.HPMax, in.HPTemp, in.ArmorClass, in.Speed, in.ProficiencyBonus,
		in.CP, in.SP, in.EP, in.GP, in.PP,
		in.Abilities.Strength, in.Abilities.Dexterity, in.Abilities.Constitution,
		in.Abilities.Intelligence, in.Abilities.Wisdom, in.Abilities.Charisma,
		in.Notes, in.DeathSaveSuccesses, in.DeathSaveFailures,
		in.HitDiceTotal, in.HitDiceUsed, in.HitDiceDie,
		in.DamageResistances, in.DamageVulnerabilities, in.DamageImmunities, in.ConditionImmunities, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CharacterRepository) Delete(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM characters WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Rest & Recovery

func (r *CharacterRepository) LongRest(ctx context.Context, id int64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Reset HP to max, clear temp HP, reset death saves, recover half hit dice.
	res, err := tx.ExecContext(ctx,
		`UPDATE characters SET
			hp_current = hp_max,
			hp_temp = 0,
			death_save_successes = 0,
			death_save_failures = 0,
			hit_dice_used = GREATEST(0, hit_dice_used - GREATEST(1, FLOOR(hit_dice_total / 2)))
		 WHERE id = ?`, id)
	if err != nil {
		return err
	}
	if err := affected(res); err != nil {
		return err
	}

	// Reset all spell slots to 0 used.
	if _, err := tx.ExecContext(ctx, `UPDATE character_spell_slots SET used = 0 WHERE character_id = ?`, id); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *CharacterRepository) ShortRest(ctx context.Context, id int64, hitDiceSpent int, hpHealed int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			hp_current = LEAST(hp_max, hp_current + ?),
			hit_dice_used = LEAST(hit_dice_total, hit_dice_used + ?)
		 WHERE id = ?`, hpHealed, hitDiceSpent, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// Items

const itemColumns = `id, character_id, name, quantity, weight, cost_cp, equipped,
	attack_bonus, damage, damage_type, description, created_at, updated_at`

func (r *CharacterRepository) Items(ctx context.Context, characterID int64) ([]model.CharacterItem, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+itemColumns+` FROM character_items WHERE character_id = ? ORDER BY equipped DESC, name`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []model.CharacterItem{}
	for rows.Next() {
		var i model.CharacterItem
		if err := rows.Scan(&i.ID, &i.CharacterID, &i.Name, &i.Quantity, &i.Weight, &i.CostCP, &i.Equipped,
			&i.AttackBonus, &i.Damage, &i.DamageType, &i.Description, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	return items, rows.Err()
}

func (r *CharacterRepository) CreateItem(ctx context.Context, characterID int64, in model.CharacterItemInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO character_items (character_id, name, quantity, weight, cost_cp, equipped,
			attack_bonus, damage, damage_type, description)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		characterID, in.Name, in.Quantity, in.Weight, in.CostCP, in.Equipped,
		in.AttackBonus, in.Damage, in.DamageType, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CharacterRepository) UpdateItem(ctx context.Context, id int64, in model.CharacterItemInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE character_items SET name = ?, quantity = ?, weight = ?, cost_cp = ?, equipped = ?,
			attack_bonus = ?, damage = ?, damage_type = ?, description = ? WHERE id = ?`,
		in.Name, in.Quantity, in.Weight, in.CostCP, in.Equipped,
		in.AttackBonus, in.Damage, in.DamageType, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CharacterRepository) DeleteItem(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM character_items WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// ItemCharacter reports which character an item belongs to, so the permission
// check reads the stored row instead of trusting the request.
func (r *CharacterRepository) ItemCharacter(ctx context.Context, id int64) (int64, error) {
	var characterID int64
	err := r.db.QueryRowContext(ctx, `SELECT character_id FROM character_items WHERE id = ?`, id).Scan(&characterID)
	return characterID, err
}

// Spells

const spellColumns = `id, character_id, name, level, school, prepared, description, created_at, updated_at`

func (r *CharacterRepository) Spells(ctx context.Context, characterID int64) ([]model.CharacterSpell, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT `+spellColumns+` FROM character_spells WHERE character_id = ? ORDER BY level, name`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	spells := []model.CharacterSpell{}
	for rows.Next() {
		var s model.CharacterSpell
		if err := rows.Scan(&s.ID, &s.CharacterID, &s.Name, &s.Level, &s.School,
			&s.Prepared, &s.Description, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		spells = append(spells, s)
	}
	return spells, rows.Err()
}

func (r *CharacterRepository) CreateSpell(ctx context.Context, characterID int64, in model.CharacterSpellInput) (int64, error) {
	res, err := r.db.ExecContext(ctx,
		`INSERT INTO character_spells (character_id, name, level, school, prepared, description)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		characterID, in.Name, in.Level, in.School, in.Prepared, in.Description)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *CharacterRepository) UpdateSpell(ctx context.Context, id int64, in model.CharacterSpellInput) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE character_spells SET name = ?, level = ?, school = ?, prepared = ?, description = ? WHERE id = ?`,
		in.Name, in.Level, in.School, in.Prepared, in.Description, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CharacterRepository) DeleteSpell(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM character_spells WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *CharacterRepository) SpellCharacter(ctx context.Context, id int64) (int64, error) {
	var characterID int64
	err := r.db.QueryRowContext(ctx, `SELECT character_id FROM character_spells WHERE id = ?`, id).Scan(&characterID)
	return characterID, err
}

// Spell slots

func (r *CharacterRepository) Slots(ctx context.Context, characterID int64) ([]model.SpellSlot, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT level, total, used FROM character_spell_slots WHERE character_id = ? ORDER BY level`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	slots := []model.SpellSlot{}
	for rows.Next() {
		var s model.SpellSlot
		if err := rows.Scan(&s.Level, &s.Total, &s.Used); err != nil {
			return nil, err
		}
		slots = append(slots, s)
	}
	return slots, rows.Err()
}

// SetSlot upserts one level, so the caller does not have to know whether a row
// exists yet.
func (r *CharacterRepository) SetSlot(ctx context.Context, characterID int64, level, total, used int) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO character_spell_slots (character_id, level, total, used) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE total = VALUES(total), used = VALUES(used)`,
		characterID, level, total, used)
	return err
}

// Item returns one inventory row on its own, which is what an attack needs.
func (r *CharacterRepository) Item(ctx context.Context, id int64) (model.CharacterItem, error) {
	var i model.CharacterItem
	err := r.db.QueryRowContext(ctx, `SELECT `+itemColumns+` FROM character_items WHERE id = ?`, id).
		Scan(&i.ID, &i.CharacterID, &i.Name, &i.Quantity, &i.Weight, &i.CostCP, &i.Equipped,
			&i.AttackBonus, &i.Damage, &i.DamageType, &i.Description, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

// Proficiencies

func (r *CharacterRepository) Proficiencies(ctx context.Context, characterID int64) ([]model.Proficiency, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT kind, name, expertise FROM character_proficiencies WHERE character_id = ? ORDER BY kind, name`, characterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Proficiency{}
	for rows.Next() {
		var p model.Proficiency
		if err := rows.Scan(&p.Kind, &p.Name, &p.Expertise); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// SetProficiency adds, updates or removes one trained skill or save.
func (r *CharacterRepository) SetProficiency(ctx context.Context, characterID int64, p model.Proficiency, trained bool) error {
	if !trained {
		_, err := r.db.ExecContext(ctx,
			`DELETE FROM character_proficiencies WHERE character_id = ? AND kind = ? AND name = ?`,
			characterID, p.Kind, p.Name)
		return err
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO character_proficiencies (character_id, kind, name, expertise) VALUES (?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE expertise = VALUES(expertise)`,
		characterID, p.Kind, p.Name, p.Expertise)
	return err
}

// ApplyDamage moves a character hit points by delta.
// If delta < 0 (damage), it absorbs hp_temp first, and remaining reduces hp_current.
// If delta > 0 (healing), it increases hp_current up to hp_max and resets death saves if revived from 0.
func (r *CharacterRepository) ApplyDamage(ctx context.Context, id int64, delta int) error {
	if delta >= 0 {
		res, err := r.db.ExecContext(ctx,
			`UPDATE characters SET
				death_save_successes = IF(hp_current = 0 AND ? > 0, 0, death_save_successes),
				death_save_failures = IF(hp_current = 0 AND ? > 0, 0, death_save_failures),
				hp_current = LEAST(hp_max, hp_current + ?)
			 WHERE id = ?`, delta, delta, delta, id)
		if err != nil {
			return err
		}
		return affected(res)
	}

	dmg := -delta
	res, err := r.db.ExecContext(ctx,
		`UPDATE characters SET
			hp_current = GREATEST(0, hp_current - GREATEST(0, ? - hp_temp)),
			hp_temp = GREATEST(0, hp_temp - ?)
		 WHERE id = ?`, dmg, dmg, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// UpdateCoins adjusts currency coins for a character.
func (r *CharacterRepository) UpdateCoins(ctx context.Context, id int64, cp, sp, ep, gp, pp int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE characters SET cp = ?, sp = ?, ep = ?, gp = ?, pp = ? WHERE id = ?`,
		cp, sp, ep, gp, pp, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// SplitPartyLoot distributes coins among all active characters in a campaign.
func (r *CharacterRepository) SplitPartyLoot(ctx context.Context, campaignID int64, cp, sp, ep, gp, pp int) error {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM characters WHERE campaign_id = ? AND status = 'active'`, campaignID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}

	shareCP := cp / len(ids)
	shareSP := sp / len(ids)
	shareEP := ep / len(ids)
	shareGP := gp / len(ids)
	sharePP := pp / len(ids)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		if _, err := tx.ExecContext(ctx,
			`UPDATE characters SET
				cp = cp + ?,
				sp = sp + ?,
				ep = ep + ?,
				gp = gp + ?,
				pp = pp + ?
			 WHERE id = ?`, shareCP, shareSP, shareEP, shareGP, sharePP, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}
