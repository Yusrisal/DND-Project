package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type MapService struct {
	repo *repository.MapRepository
}

func NewMapService(repo *repository.MapRepository) *MapService {
	return &MapService{repo: repo}
}

func (s *MapService) List(ctx context.Context, campaignID int64, q model.Query) (model.Page[model.Map], error) {
	items, total, err := s.repo.List(ctx, campaignID, q)
	if err != nil {
		return model.Page[model.Map]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *MapService) Get(ctx context.Context, id int64) (model.Map, error) {
	m, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return m, ErrNotFound
	}
	return m, err
}

func (s *MapService) Create(ctx context.Context, in model.MapInput) (model.Map, error) {
	in, err := normalizeMap(in)
	if err != nil {
		return model.Map{}, err
	}
	if in.CampaignID <= 0 {
		return model.Map{}, invalid("campaign_id", "is required")
	}
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Map{}, err
	}
	return s.Get(ctx, id)
}

func (s *MapService) Update(ctx context.Context, id int64, in model.MapInput) (model.Map, error) {
	in, err := normalizeMap(in)
	if err != nil {
		return model.Map{}, err
	}
	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Map{}, ErrNotFound
		}
		return model.Map{}, err
	}
	return s.Get(ctx, id)
}

func (s *MapService) UpdateFog(ctx context.Context, id int64, enabled bool, revealed string) (model.Map, error) {
	if err := s.repo.UpdateFog(ctx, id, enabled, revealed); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Map{}, ErrNotFound
		}
		return model.Map{}, err
	}
	return s.Get(ctx, id)
}

func (s *MapService) SetWeather(ctx context.Context, id int64, weather string) (model.Map, error) {
	weather = strings.TrimSpace(weather)
	if weather == "" {
		weather = "clear"
	}
	if err := s.repo.SetWeather(ctx, id, weather); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Map{}, ErrNotFound
		}
		return model.Map{}, err
	}
	return s.Get(ctx, id)
}

func (s *MapService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MapService) ListAoETemplates(ctx context.Context, mapID int64) ([]model.AoETemplate, error) {
	return s.repo.ListAoETemplates(ctx, mapID)
}

func (s *MapService) CreateAoETemplate(ctx context.Context, in model.AoETemplateInput) (model.AoETemplate, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = "Spell Template"
	}
	if in.SizeFt <= 0 {
		in.SizeFt = 20
	}
	if in.Color == "" {
		in.Color = "#e74c3c"
	}
	return s.repo.CreateAoETemplate(ctx, in)
}

func (s *MapService) DeleteAoETemplate(ctx context.Context, id int64) error {
	err := s.repo.DeleteAoETemplate(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Map Objects (Doors, Chests, Traps, Markers)

func (s *MapService) ListObjects(ctx context.Context, mapID int64) ([]model.MapObject, error) {
	return s.repo.ListObjects(ctx, mapID)
}

func (s *MapService) CreateObject(ctx context.Context, in model.MapObjectInput) (model.MapObject, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = in.Kind
	}
	if in.State == "" {
		in.State = "closed"
	}
	return s.repo.CreateObject(ctx, in)
}

func (s *MapService) UpdateObject(ctx context.Context, id int64, in model.MapObjectInput) (model.MapObject, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = in.Kind
	}
	return s.repo.UpdateObject(ctx, id, in)
}

func (s *MapService) DeleteObject(ctx context.Context, id int64) error {
	err := s.repo.DeleteObject(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func normalizeMap(in model.MapInput) (model.MapInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.ImageURL = strings.TrimSpace(in.ImageURL)
	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.GridSize <= 0 {
		in.GridSize = 50
	}
	if in.Width <= 0 {
		in.Width = 1920
	}
	if in.Height <= 0 {
		in.Height = 1080
	}
	if in.Weather == "" {
		in.Weather = "clear"
	}
	return in, nil
}
