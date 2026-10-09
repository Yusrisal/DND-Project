package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type TokenService struct {
	repo *repository.TokenRepository
}

func NewTokenService(repo *repository.TokenRepository) *TokenService {
	return &TokenService{repo: repo}
}

// Board returns every token on a map, which is what a client loads on join.
// Conditions are attached in one extra query rather than one per token.
func (s *TokenService) Board(ctx context.Context, mapID int64) ([]model.Token, error) {
	tokens, err := s.repo.ListByMap(ctx, mapID)
	if err != nil {
		return nil, err
	}
	return tokens, s.repo.LoadConditions(ctx, tokens)
}

func (s *TokenService) List(ctx context.Context, mapID int64, q model.Query) (model.Page[model.Token], error) {
	items, total, err := s.repo.List(ctx, mapID, q)
	if err != nil {
		return model.Page[model.Token]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *TokenService) Get(ctx context.Context, id int64) (model.Token, error) {
	t, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return t, ErrNotFound
	}
	if err != nil {
		return t, err
	}

	list := []model.Token{t}
	if err := s.repo.LoadConditions(ctx, list); err != nil {
		return t, err
	}
	return list[0], nil
}

func (s *TokenService) Create(ctx context.Context, in model.TokenInput) (model.Token, error) {
	in, err := normalizeToken(in)
	if err != nil {
		return model.Token{}, err
	}
	if in.MapID <= 0 {
		return model.Token{}, invalid("map_id", "is required")
	}
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Token{}, err
	}
	return s.Get(ctx, id)
}

func (s *TokenService) Update(ctx context.Context, id int64, in model.TokenInput) (model.Token, error) {
	in, err := normalizeToken(in)
	if err != nil {
		return model.Token{}, err
	}
	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Token{}, ErrNotFound
		}
		return model.Token{}, err
	}
	return s.Get(ctx, id)
}

// Move persists a drag. A player may only move a token they own; the DM may
// move anything, including unowned monsters. The ownership check reads the
// stored row rather than trusting the client, and the token it returns carries
// the real map id so the caller broadcasts to the right room.
func (s *TokenService) Move(ctx context.Context, user model.User, move model.TokenMove) (model.Token, error) {
	token, err := s.Get(ctx, move.ID)
	if err != nil {
		return model.Token{}, err
	}
	if !user.IsDM() && (token.OwnerUserID == nil || *token.OwnerUserID != user.ID) {
		return model.Token{}, ErrForbidden
	}

	if err := s.repo.Move(ctx, move.ID, move.X, move.Y); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Token{}, ErrNotFound
		}
		return model.Token{}, err
	}
	return s.Get(ctx, move.ID)
}

func (s *TokenService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeToken(in model.TokenInput) (model.TokenInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	in.Color = strings.TrimSpace(in.Color)
	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Color == "" {
		in.Color = "#c0392b"
	}
	if in.Size <= 0 {
		in.Size = 1
	}
	return in, nil
}
