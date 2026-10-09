package model

import "time"

// ChatMessage keeps author_name alongside user_id so history stays readable
// after an account is deleted, which sets user_id to NULL.
type ChatMessage struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	UserID     *int64    `json:"user_id"`
	AuthorName string    `json:"author_name"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
}

type ChatMessageInput struct {
	Body string `json:"body"`
}
