package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type HandoutService struct {
	repo *repository.HandoutRepository
}

func NewHandoutService(repo *repository.HandoutRepository) *HandoutService {
	return &HandoutService{repo: repo}
}

func (s *HandoutService) List(ctx context.Context, campaignID int64, isDM bool, q model.Query) (model.Page[model.Handout], error) {
	revealedOnly := !isDM
	items, total, err := s.repo.List(ctx, campaignID, revealedOnly, q)
	if err != nil {
		return model.Page[model.Handout]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *HandoutService) Get(ctx context.Context, id int64, isDM bool) (model.Handout, error) {
	h, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return h, ErrNotFound
	}
	if err != nil {
		return h, err
	}
	if !isDM && !h.IsRevealed {
		return model.Handout{}, ErrNotFound
	}
	return h, nil
}

func (s *HandoutService) Create(ctx context.Context, in model.HandoutInput) (model.Handout, error) {
	in, err := normalizeHandout(in)
	if err != nil {
		return model.Handout{}, err
	}
	if in.CampaignID <= 0 {
		return model.Handout{}, invalid("campaign_id", "is required")
	}
	return s.repo.Create(ctx, in)
}

func (s *HandoutService) Update(ctx context.Context, id int64, in model.HandoutInput) (model.Handout, error) {
	in, err := normalizeHandout(in)
	if err != nil {
		return model.Handout{}, err
	}
	h, err := s.repo.Update(ctx, id, in)
	if errors.Is(err, sql.ErrNoRows) {
		return h, ErrNotFound
	}
	return h, err
}

func (s *HandoutService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeHandout(in model.HandoutInput) (model.HandoutInput, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	if in.Title == "" {
		return in, invalid("title", "is required")
	}
	return in, nil
}
