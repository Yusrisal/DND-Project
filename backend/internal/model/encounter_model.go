package model

import "time"

// Encounter is the initiative order for one map. One is active at a time; the
// DM ends it rather than deleting it, so the log of the fight survives.
type Encounter struct {
	ID         int64       `json:"id"`
	MapID      int64       `json:"map_id"`
	Round      int         `json:"round"`
	TurnIndex  int         `json:"turn_index"`
	Active     bool        `json:"active"`
	Combatants []Combatant `json:"combatants"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// Current returns whose turn it is, or false when the encounter is empty.
func (e Encounter) Current() (Combatant, bool) {
	if e.TurnIndex < 0 || e.TurnIndex >= len(e.Combatants) {
		return Combatant{}, false
	}
	return e.Combatants[e.TurnIndex], true
}

// Combatant is one line in the initiative order. TokenID is optional so the DM
// can track something that is not on the board, such as a lair action.
type Combatant struct {
	ID         int64  `json:"id"`
	TokenID    *int64 `json:"token_id"`
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
}

type CombatantInput struct {
	TokenID    *int64 `json:"token_id"`
	Name       string `json:"name"`
	Initiative int    `json:"initiative"`
}

// Vitals is what an attack needs to know about a target, resolved from the
// linked character when there is one and from the token itself otherwise.
type Vitals struct {
	TokenID               int64  `json:"token_id"`
	CharacterID           *int64 `json:"character_id"`
	Name                  string `json:"name"`
	HPCurrent             int    `json:"hp_current"`
	HPMax                 int    `json:"hp_max"`
	HPTemp                int    `json:"hp_temp"`
	ArmorClass            int    `json:"armor_class"`
	DamageResistances     string `json:"damage_resistances"`
	DamageVulnerabilities string `json:"damage_vulnerabilities"`
	DamageImmunities      string `json:"damage_immunities"`
	ConditionImmunities   string `json:"condition_immunities"`
}

// AttackRequest names one attack and a target. Exactly one source is given:
// ItemID for a character weapon, which requires owning that sheet, or ActionID
// plus AttackerTokenID for a monster action, which requires being the DM.
type AttackRequest struct {
	ItemID          int64 `json:"item_id"`
	ActionID        int64 `json:"action_id"`
	AttackerTokenID int64 `json:"attacker_token_id"`
	TargetTokenID   int64 `json:"target_token_id"`
	Advantage       bool  `json:"advantage"`
	Disadvantage    bool  `json:"disadvantage"`
}

// AttackResult is the whole resolution, broadcast so the table can check it.
type AttackResult struct {
	AttackerName string       `json:"attacker_name"`
	TargetName   string       `json:"target_name"`
	WeaponName   string       `json:"weapon_name"`
	AttackRoll   DiceSummary  `json:"attack_roll"`
	TargetAC     int          `json:"target_ac"`
	Hit          bool         `json:"hit"`
	Critical     bool         `json:"critical"`
	Fumble       bool         `json:"fumble"`
	Damage       *DiceSummary `json:"damage"`
	DamageType   string       `json:"damage_type"`
	Target       Vitals       `json:"target"`
}

// DiceSummary mirrors the roll breakdown without importing the service package
// into the model layer.
type DiceSummary struct {
	Expression string `json:"expression"`
	Total      int    `json:"total"`
	Detail     any    `json:"detail"`
}
