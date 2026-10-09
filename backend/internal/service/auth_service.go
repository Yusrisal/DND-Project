package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// SessionTTL is how long a login lasts. Sessions are revocable, so this can be
// generous without the usual token tradeoff.
const SessionTTL = 30 * 24 * time.Hour

const minPasswordLength = 10

type AuthService struct {
	users    *repository.UserRepository
	sessions *repository.SessionRepository
}

func NewAuthService(users *repository.UserRepository, sessions *repository.SessionRepository) *AuthService {
	return &AuthService{users: users, sessions: sessions}
}

// Register creates an account against an invite code. The first account on a
// fresh database needs no code and becomes the DM.
func (s *AuthService) Register(ctx context.Context, in model.RegisterInput) (model.User, error) {
	email, err := normalizeEmail(in.Email)
	if err != nil {
		return model.User{}, err
	}
	name := strings.TrimSpace(in.DisplayName)
	if name == "" {
		return model.User{}, invalid("display_name", "is required")
	}
	if len(name) > 80 {
		return model.User{}, invalid("display_name", "must be at most 80 characters")
	}
	if len(in.Password) < minPasswordLength {
		return model.User{}, invalid("password", "must be at least 10 characters")
	}

	// Checked up front for a clear message; the unique index is the real guard.
	taken, err := s.users.EmailTaken(ctx, email)
	if err != nil {
		return model.User{}, err
	}
	if taken {
		return model.User{}, invalid("email", "is already registered")
	}

	hash, err := hashPassword(in.Password)
	if err != nil {
		return model.User{}, err
	}

	user, err := s.users.CreateWithInvite(ctx, strings.TrimSpace(in.Code), email, name, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, invalid("code", "is invalid, expired or already used")
	}
	return user, err
}

// Login returns the raw session token, which the caller puts in a cookie. Only
// its hash is stored, so a database leak does not hand over live sessions.
func (s *AuthService) Login(ctx context.Context, c model.Credentials) (model.User, string, time.Time, error) {
	email, err := normalizeEmail(c.Email)
	if err != nil {
		return model.User{}, "", time.Time{}, ErrUnauthorized
	}

	user, hash, err := s.users.GetCredentials(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		// Hash anyway, so a missing account and a wrong password take the same time.
		_, _ = hashPassword(c.Password)
		return model.User{}, "", time.Time{}, ErrUnauthorized
	}
	if err != nil {
		return model.User{}, "", time.Time{}, err
	}

	ok, err := verifyPassword(c.Password, hash)
	if err != nil || !ok {
		return model.User{}, "", time.Time{}, ErrUnauthorized
	}

	token, err := randomToken()
	if err != nil {
		return model.User{}, "", time.Time{}, err
	}
	expiresAt := time.Now().Add(SessionTTL)
	if err := s.sessions.Create(ctx, hashToken(token), user.ID, expiresAt); err != nil {
		return model.User{}, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessions.Delete(ctx, hashToken(token))
}

// UserFromToken resolves a session cookie. An expired session is filtered out
// by the query, so it reads the same as no session at all.
func (s *AuthService) UserFromToken(ctx context.Context, token string) (model.User, error) {
	if token == "" {
		return model.User{}, ErrUnauthorized
	}
	user, err := s.sessions.User(ctx, hashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return model.User{}, ErrUnauthorized
	}
	return user, err
}

func (s *AuthService) ListUsers(ctx context.Context, q model.Query) (model.Page[model.User], error) {
	items, total, err := s.users.List(ctx, q)
	if err != nil {
		return model.Page[model.User]{}, err
	}
	return model.NewPage(items, total, q), nil
}

// PurgeExpiredSessions is called on a timer by the server.
func (s *AuthService) PurgeExpiredSessions(ctx context.Context) (int64, error) {
	return s.sessions.DeleteExpired(ctx)
}

func normalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", invalid("email", "is required")
	}
	at := strings.Index(email, "@")
	if at < 1 || at == len(email)-1 || len(email) > 191 {
		return "", invalid("email", "is not a valid address")
	}
	return email, nil
}

func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// hashToken is a plain SHA-256: the token is already 256 bits of entropy, so
// there is nothing for a slow KDF to protect against here.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
