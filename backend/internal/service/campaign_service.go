package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type CampaignService struct {
	repo *repository.CampaignRepository
}

func NewCampaignService(repo *repository.CampaignRepository) *CampaignService {
	return &CampaignService{repo: repo}
}

func (s *CampaignService) List(ctx context.Context, q model.Query) (model.Page[model.Campaign], error) {
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return model.Page[model.Campaign]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *CampaignService) Get(ctx context.Context, id int64) (model.Campaign, error) {
	c, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return c, ErrNotFound
	}
	return c, err
}

func (s *CampaignService) Create(ctx context.Context, in model.CampaignInput) (model.Campaign, error) {
	in, err := normalizeCampaign(in)
	if err != nil {
		return model.Campaign{}, err
	}
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Campaign{}, err
	}
	return s.Get(ctx, id)
}

func (s *CampaignService) Update(ctx context.Context, id int64, in model.CampaignInput) (model.Campaign, error) {
	in, err := normalizeCampaign(in)
	if err != nil {
		return model.Campaign{}, err
	}
	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Campaign{}, ErrNotFound
		}
		return model.Campaign{}, err
	}
	return s.Get(ctx, id)
}

func (s *CampaignService) SetPause(ctx context.Context, id int64, in model.PauseInput) (model.Campaign, error) {
	msg := strings.TrimSpace(in.PauseMessage)
	if msg == "" {
		msg = "Session Paused by DM"
	}
	if err := s.repo.SetPause(ctx, id, in.IsPaused, msg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Campaign{}, ErrNotFound
		}
		return model.Campaign{}, err
	}
	return s.Get(ctx, id)
}

func (s *CampaignService) SetRetention(ctx context.Context, id int64, days int) (model.Campaign, error) {
	if days < 0 {
		days = 0
	}
	if err := s.repo.SetRetention(ctx, id, days); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Campaign{}, ErrNotFound
		}
		return model.Campaign{}, err
	}
	return s.Get(ctx, id)
}

func (s *CampaignService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeCampaign(in model.CampaignInput) (model.CampaignInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	return in, nil
}
