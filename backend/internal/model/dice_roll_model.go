package model

import (
	"encoding/json"
	"time"
)

// DiceRoll is the record of a roll. Detail holds the per-die breakdown as it
// was produced by the server, so a result can be checked afterwards.
type DiceRoll struct {
	ID          int64           `json:"id"`
	CampaignID  int64           `json:"campaign_id"`
	UserID      *int64          `json:"user_id"`
	CharacterID *int64          `json:"character_id"`
	AuthorName  string          `json:"author_name"`
	Label       string          `json:"label"`
	Expression  string          `json:"expression"`
	Detail      json.RawMessage `json:"detail"`
	Total       int             `json:"total"`
	CreatedAt   time.Time       `json:"created_at"`
}

// RollRequest is what a client sends. It never carries a result.
type RollRequest struct {
	Expression  string `json:"expression"`
	Label       string `json:"label"`
	CharacterID *int64 `json:"character_id"`
}
