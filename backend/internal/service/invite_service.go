package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"time"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// InviteTTL is how long a freshly minted code stays usable.
const InviteTTL = 7 * 24 * time.Hour

// codeAlphabet leaves out the characters that get misread when a code is typed
// from a chat message: 0/O, 1/I/l.
const codeAlphabet = "23456789abcdefghjkmnpqrstuvwxyz"

const codeLength = 10

type InviteService struct {
	invites *repository.InviteRepository
}

func NewInviteService(invites *repository.InviteRepository) *InviteService {
	return &InviteService{invites: invites}
}

func (s *InviteService) Create(ctx context.Context, createdBy int64, in model.InviteInput) (model.Invite, error) {
	role := in.Role
	if role == "" {
		role = model.RolePlayer
	}
	if !role.Valid() {
		return model.Invite{}, invalid("role", "must be dm or player")
	}

	code, err := randomCode()
	if err != nil {
		return model.Invite{}, err
	}
	if err := s.invites.Create(ctx, code, role, createdBy, time.Now().Add(InviteTTL)); err != nil {
		return model.Invite{}, err
	}
	return s.invites.Get(ctx, code)
}

func (s *InviteService) List(ctx context.Context, q model.Query) (model.Page[model.Invite], error) {
	items, total, err := s.invites.List(ctx, q)
	if err != nil {
		return model.Page[model.Invite]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *InviteService) Delete(ctx context.Context, code string) error {
	err := s.invites.Delete(ctx, code)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func randomCode() (string, error) {
	buf := make([]byte, codeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	// Modulo bias is irrelevant here: the alphabet is 31 characters and the code
	// only has to be unguessable for a week, not uniformly distributed.
	out := make([]byte, codeLength)
	for i, b := range buf {
		out[i] = codeAlphabet[int(b)%len(codeAlphabet)]
	}
	return string(out), nil
}
