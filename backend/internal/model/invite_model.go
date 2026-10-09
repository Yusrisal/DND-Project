package model

import "time"

type Invite struct {
	Code      string     `json:"code"`
	Role      Role       `json:"role"`
	CreatedBy int64      `json:"created_by"`
	UsedBy    *int64     `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

type InviteInput struct {
	Role Role `json:"role"`
}
