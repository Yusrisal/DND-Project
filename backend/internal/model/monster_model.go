package model

import "time"

// Monster is a bestiary entry: the template a token is spawned from. The token
// is the instance that takes damage, so several goblins share one stat block
// without sharing hit points.
type Monster struct {
	ID                    int64     `json:"id"`
	Name                  string    `json:"name"`
	Size                  string    `json:"size"`
	Type                  string    `json:"type"`
	Alignment             string    `json:"alignment"`
	ArmorClass            int       `json:"armor_class"`
	HPAverage             int       `json:"hp_average"`
	HPFormula             string    `json:"hp_formula"`
	Speed                 int       `json:"speed"`
	Abilities             Abilities `json:"abilities"`
	Modifiers             Abilities `json:"modifiers"`
	ChallengeRating       string    `json:"challenge_rating"`
	Senses                string    `json:"senses"`
	Languages             string    `json:"languages"`
	Traits                string    `json:"traits"`
	DamageResistances     string    `json:"damage_resistances"`
	DamageVulnerabilities string    `json:"damage_vulnerabilities"`
	DamageImmunities      string    `json:"damage_immunities"`
	ConditionImmunities   string    `json:"condition_immunities"`
	Color                 string    `json:"color"`
	ImageURL              string    `json:"image_url"`
	TokenSize             int       `json:"token_size"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type MonsterInput struct {
	Name                  string    `json:"name"`
	Size                  string    `json:"size"`
	Type                  string    `json:"type"`
	Alignment             string    `json:"alignment"`
	ArmorClass            int       `json:"armor_class"`
	HPAverage             int       `json:"hp_average"`
	HPFormula             string    `json:"hp_formula"`
	Speed                 int       `json:"speed"`
	Abilities             Abilities `json:"abilities"`
	ChallengeRating       string    `json:"challenge_rating"`
	Senses                string    `json:"senses"`
	Languages             string    `json:"languages"`
	Traits                string    `json:"traits"`
	DamageResistances     string    `json:"damage_resistances"`
	DamageVulnerabilities string    `json:"damage_vulnerabilities"`
	DamageImmunities      string    `json:"damage_immunities"`
	ConditionImmunities   string    `json:"condition_immunities"`
	Color                 string    `json:"color"`
	ImageURL              string    `json:"image_url"`
	TokenSize             int       `json:"token_size"`
}

// MonsterProficiency is a flat modifier, the way a stat block prints a skill:
// "Perception +4". There is no proficiency bonus to apply on top.
type MonsterProficiency struct {
	Kind     string `json:"kind"`
	Name     string `json:"name"`
	Modifier int    `json:"modifier"`
}

// MonsterAction is an attack when it carries a damage expression, and a trait
// the DM reads out when it does not.
type MonsterAction struct {
	ID          int64  `json:"id"`
	MonsterID   int64  `json:"monster_id"`
	Name        string `json:"name"`
	AttackBonus int    `json:"attack_bonus"`
	Damage      string `json:"damage"`
	DamageType  string `json:"damage_type"`
	Description string `json:"description"`
}

func (a MonsterAction) IsAttack() bool { return a.Damage != "" }

type MonsterActionInput struct {
	Name        string `json:"name"`
	AttackBonus int    `json:"attack_bonus"`
	Damage      string `json:"damage"`
	DamageType  string `json:"damage_type"`
	Description string `json:"description"`
}

// MonsterStatBlock is the whole entry, which is what the bestiary panel and the
// DM view of a spawned token both load.
type MonsterStatBlock struct {
	Monster       Monster              `json:"monster"`
	Proficiencies []MonsterProficiency `json:"proficiencies"`
	Actions       []MonsterAction      `json:"actions"`
}

// SpawnInput places a monster on a map. Count spawns several at once, numbered
// so the initiative order can tell them apart.
type SpawnInput struct {
	MonsterID int64   `json:"monster_id"`
	MapID     int64   `json:"map_id"`
	Count     int     `json:"count"`
	RollHP    bool    `json:"roll_hp"`
	X         float64 `json:"x"`
	Y         float64 `json:"y"`
}
