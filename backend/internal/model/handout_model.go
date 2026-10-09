package model

import "time"

type Handout struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	ImageURL   string    `json:"image_url"`
	IsRevealed bool      `json:"is_revealed"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HandoutInput struct {
	CampaignID int64  `json:"campaign_id"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	ImageURL   string `json:"image_url"`
	IsRevealed bool   `json:"is_revealed"`
}
