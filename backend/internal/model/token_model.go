package model

import "time"

// Token is a figure on the board. CharacterID links it to a sheet, in which
// case the sheet owns the hit points and the armour class and the columns here
// are ignored; an unlinked token, such as a monster, carries its own.
type Token struct {
	ID          int64     `json:"id"`
	MapID       int64     `json:"map_id"`
	OwnerUserID *int64    `json:"owner_user_id"`
	CharacterID *int64    `json:"character_id"`
	MonsterID   *int64    `json:"monster_id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	Color       string    `json:"color"`
	Size        int       `json:"size"`
	HPCurrent   int       `json:"hp_current"`
	HPMax       int       `json:"hp_max"`
	ArmorClass  int       `json:"armor_class"`
	Conditions  []string  `json:"conditions"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TokenInput struct {
	MapID       int64   `json:"map_id"`
	OwnerUserID *int64  `json:"owner_user_id"`
	CharacterID *int64  `json:"character_id"`
	MonsterID   *int64  `json:"monster_id"`
	Name        string  `json:"name"`
	ImageURL    string  `json:"image_url"`
	Color       string  `json:"color"`
	Size        int     `json:"size"`
	HPCurrent   int     `json:"hp_current"`
	HPMax       int     `json:"hp_max"`
	ArmorClass  int     `json:"armor_class"`
	X           float64 `json:"x"`
	Y           float64 `json:"y"`
}

type TokenMove struct {
	ID int64   `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
}

type TokenCondition struct {
	TokenID         int64  `json:"token_id"`
	ConditionName   string `json:"condition_name"`
	DurationRounds  int    `json:"duration_rounds"`
	RemainingRounds int    `json:"remaining_rounds"`
	Notes           string `json:"notes"`
}

// ConditionInput toggles one status on a token with optional duration and notes.
type ConditionInput struct {
	Condition      string `json:"condition"`
	Active         bool   `json:"active"`
	DurationRounds int    `json:"duration_rounds"`
	Notes          string `json:"notes"`
}
