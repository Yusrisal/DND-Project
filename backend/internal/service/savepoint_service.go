package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type SavepointService struct {
	repo       *repository.SavepointRepository
	tokens     *repository.TokenRepository
	characters *repository.CharacterRepository
	encounters *repository.EncounterRepository
}

func NewSavepointService(
	repo *repository.SavepointRepository,
	tokens *repository.TokenRepository,
	characters *repository.CharacterRepository,
	encounters *repository.EncounterRepository,
) *SavepointService {
	return &SavepointService{
		repo:       repo,
		tokens:     tokens,
		characters: characters,
		encounters: encounters,
	}
}

func (s *SavepointService) List(ctx context.Context, campaignID int64, q model.Query) (model.Page[model.Savepoint], error) {
	items, total, err := s.repo.List(ctx, campaignID, q)
	if err != nil {
		return model.Page[model.Savepoint]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *SavepointService) Get(ctx context.Context, id int64) (model.Savepoint, error) {
	sp, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return sp, ErrNotFound
	}
	return sp, err
}

func (s *SavepointService) Create(ctx context.Context, in model.SavepointInput) (model.Savepoint, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		in.Name = "Checkpoint " + time.Now().Format("Jan 02 15:04")
	}
	if in.CampaignID <= 0 {
		return model.Savepoint{}, invalid("campaign_id", "is required")
	}

	snapshot := model.SavepointSnapshot{
		CapturedAt: time.Now(),
		MapID:      in.MapID,
	}

	if in.MapID != nil && *in.MapID > 0 {
		tokens, err := s.tokens.ListByMap(ctx, *in.MapID)
		if err == nil {
			snapshot.Tokens = tokens
		}
		if enc, err := s.encounters.Active(ctx, *in.MapID); err == nil {
			snapshot.Encounter = &enc
		}
	}

	chars, _, err := s.characters.List(ctx, in.CampaignID, 0, model.Query{PageSize: 100})
	if err == nil {
		snapshot.Characters = chars
	}

	raw, err := json.Marshal(snapshot)
	if err != nil {
		return model.Savepoint{}, err
	}

	return s.repo.Create(ctx, in.CampaignID, in.MapID, in.Name, in.Notes, string(raw))
}

func (s *SavepointService) Restore(ctx context.Context, id int64) (model.Savepoint, error) {
	sp, err := s.Get(ctx, id)
	if err != nil {
		return sp, err
	}

	var snapshot model.SavepointSnapshot
	if err := json.Unmarshal([]byte(sp.SnapshotData), &snapshot); err != nil {
		return sp, err
	}

	// Restore tokens positions and HP
	for _, t := range snapshot.Tokens {
		_ = s.tokens.Update(ctx, t.ID, model.TokenInput{
			OwnerUserID: t.OwnerUserID,
			CharacterID: t.CharacterID,
			MonsterID:   t.MonsterID,
			HPCurrent:   t.HPCurrent,
			HPMax:       t.HPMax,
			ArmorClass:  t.ArmorClass,
			Name:        t.Name,
			ImageURL:    t.ImageURL,
			Color:       t.Color,
			Size:        t.Size,
			X:           t.X,
			Y:           t.Y,
		})
	}

	// Restore character HP and Vitals
	for _, c := range snapshot.Characters {
		_ = s.characters.Update(ctx, c.ID, model.CharacterInput{
			CampaignID:         c.CampaignID,
			UserID:             c.UserID,
			Status:             c.Status,
			Name:               c.Name,
			Class:              c.Class,
			Race:               c.Race,
			Level:              c.Level,
			HPCurrent:          c.HPCurrent,
			HPMax:              c.HPMax,
			HPTemp:             c.HPTemp,
			ArmorClass:         c.ArmorClass,
			Speed:              c.Speed,
			ProficiencyBonus:   c.ProficiencyBonus,
			CP:                 c.CP,
			SP:                 c.SP,
			EP:                 c.EP,
			GP:                 c.GP,
			PP:                 c.PP,
			Abilities:          c.Abilities,
			Notes:              c.Notes,
			DeathSaveSuccesses: c.DeathSaveSuccesses,
			DeathSaveFailures:  c.DeathSaveFailures,
			HitDiceTotal:       c.HitDiceTotal,
			HitDiceUsed:        c.HitDiceUsed,
			HitDiceDie:         c.HitDiceDie,
		})
	}

	return sp, nil
}

func (s *SavepointService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
