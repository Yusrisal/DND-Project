package model

import "time"

type Role string

const (
	RoleDM     Role = "dm"
	RolePlayer Role = "player"
)

func (r Role) Valid() bool { return r == RoleDM || r == RolePlayer }

type User struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	Role        Role      `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (u User) IsDM() bool { return u.Role == RoleDM }

// Credentials is the login body. The password never appears on User, which is
// what every response is built from.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterInput struct {
	Code        string `json:"code"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

// Session is what a cookie resolves to. The raw token is returned once, at
// login; only its hash is stored.
type Session struct {
	UserID    int64
	ExpiresAt time.Time
}
