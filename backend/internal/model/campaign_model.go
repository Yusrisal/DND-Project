package model

import "time"

type Campaign struct {
	ID                int64     `json:"id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	IsPaused          bool      `json:"is_paused"`
	PauseMessage      string    `json:"pause_message"`
	ChatRetentionDays int       `json:"chat_retention_days"` // 0 = all time, 1 = 24h, 7 = 1 week, 14 = 2 weeks, 30 = 1 month
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type CampaignInput struct {
	Name              string `json:"name"`
	Description       string `json:"description"`
	IsPaused          bool   `json:"is_paused"`
	PauseMessage      string `json:"pause_message"`
	ChatRetentionDays int    `json:"chat_retention_days"`
}

type PauseInput struct {
	IsPaused     bool   `json:"is_paused"`
	PauseMessage string `json:"pause_message"`
}

type RetentionInput struct {
	ChatRetentionDays int `json:"chat_retention_days"`
}
