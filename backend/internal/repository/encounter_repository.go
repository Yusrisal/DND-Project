package repository

import (
	"context"
	"database/sql"

	"dndproject/backend/internal/model"
)

type EncounterRepository struct {
	db *sql.DB
}

func NewEncounterRepository(db *sql.DB) *EncounterRepository {
	return &EncounterRepository{db: db}
}

// Active returns the running encounter for a map, combatants included, ordered
// highest initiative first with the id as the tiebreak so the order is stable
// across reads.
func (r *EncounterRepository) Active(ctx context.Context, mapID int64) (model.Encounter, error) {
	var e model.Encounter
	err := r.db.QueryRowContext(ctx,
		`SELECT id, map_id, round, turn_index, active, created_at, updated_at
		   FROM encounters WHERE map_id = ? AND active = 1 ORDER BY id DESC LIMIT 1`, mapID,
	).Scan(&e.ID, &e.MapID, &e.Round, &e.TurnIndex, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return model.Encounter{}, err
	}

	e.Combatants, err = r.Combatants(ctx, e.ID)
	return e, err
}

func (r *EncounterRepository) Get(ctx context.Context, id int64) (model.Encounter, error) {
	var e model.Encounter
	err := r.db.QueryRowContext(ctx,
		`SELECT id, map_id, round, turn_index, active, created_at, updated_at FROM encounters WHERE id = ?`, id,
	).Scan(&e.ID, &e.MapID, &e.Round, &e.TurnIndex, &e.Active, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return model.Encounter{}, err
	}

	e.Combatants, err = r.Combatants(ctx, e.ID)
	return e, err
}

func (r *EncounterRepository) Combatants(ctx context.Context, encounterID int64) ([]model.Combatant, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, token_id, name, initiative FROM combatants
		  WHERE encounter_id = ? ORDER BY initiative DESC, id`, encounterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []model.Combatant{}
	for rows.Next() {
		var c model.Combatant
		if err := rows.Scan(&c.ID, &c.TokenID, &c.Name, &c.Initiative); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

func (r *EncounterRepository) Create(ctx context.Context, mapID int64) (int64, error) {
	res, err := r.db.ExecContext(ctx, `INSERT INTO encounters (map_id) VALUES (?)`, mapID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *EncounterRepository) SetTurn(ctx context.Context, id int64, round, turnIndex int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE encounters SET round = ?, turn_index = ? WHERE id = ?`, round, turnIndex, id)
	if err != nil {
		return err
	}
	return affected(res)
}

// End leaves the row in place so the fight stays in the record; only the flag
// changes.
func (r *EncounterRepository) End(ctx context.Context, id int64) error {
	res, err := r.db.ExecContext(ctx, `UPDATE encounters SET active = 0 WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *EncounterRepository) AddCombatant(ctx context.Context, encounterID int64, in model.CombatantInput) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO combatants (encounter_id, token_id, name, initiative) VALUES (?, ?, ?, ?)`,
		encounterID, in.TokenID, in.Name, in.Initiative)
	return err
}

func (r *EncounterRepository) SetInitiative(ctx context.Context, combatantID int64, initiative int) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE combatants SET initiative = ? WHERE id = ?`, initiative, combatantID)
	if err != nil {
		return err
	}
	return affected(res)
}

func (r *EncounterRepository) RemoveCombatant(ctx context.Context, combatantID int64) (int64, error) {
	var encounterID int64
	if err := r.db.QueryRowContext(ctx,
		`SELECT encounter_id FROM combatants WHERE id = ?`, combatantID).Scan(&encounterID); err != nil {
		return 0, err
	}

	res, err := r.db.ExecContext(ctx, `DELETE FROM combatants WHERE id = ?`, combatantID)
	if err != nil {
		return 0, err
	}
	return encounterID, affected(res)
}
