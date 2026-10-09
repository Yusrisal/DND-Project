package model

import "time"

type Savepoint struct {
	ID           int64     `json:"id"`
	CampaignID   int64     `json:"campaign_id"`
	MapID        *int64    `json:"map_id"`
	Name         string    `json:"name"`
	Notes        string    `json:"notes"`
	SnapshotData string    `json:"snapshot_data"`
	CreatedAt    time.Time `json:"created_at"`
}

type SavepointInput struct {
	CampaignID int64  `json:"campaign_id"`
	MapID      *int64 `json:"map_id"`
	Name       string `json:"name"`
	Notes      string `json:"notes"`
}

type SavepointSnapshot struct {
	CapturedAt time.Time   `json:"captured_at"`
	MapID      *int64      `json:"map_id"`
	Tokens     []Token     `json:"tokens"`
	Characters []Character `json:"characters"`
	Encounter  *Encounter  `json:"encounter"`
}
