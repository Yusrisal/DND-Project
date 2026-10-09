package service

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type CompendiumService struct {
	repo     *repository.CompendiumRepository
	monsters *repository.MonsterRepository
}

func NewCompendiumService(repo *repository.CompendiumRepository, monsters *repository.MonsterRepository) *CompendiumService {
	return &CompendiumService{repo: repo, monsters: monsters}
}

// Environments

func (s *CompendiumService) ListEnvironments(ctx context.Context, campaignID, mapID int64, q model.Query) (model.Page[model.Environment], error) {
	items, total, err := s.repo.ListEnvironments(ctx, campaignID, mapID, q)
	if err != nil {
		return model.Page[model.Environment]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *CompendiumService) GetEnvironment(ctx context.Context, id int64) (model.Environment, error) {
	env, err := s.repo.GetEnvironment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Environment{}, ErrNotFound
	}
	return env, err
}

func (s *CompendiumService) CreateEnvironment(ctx context.Context, user model.User, in model.EnvironmentInput) (model.Environment, error) {
	if !user.IsDM() {
		return model.Environment{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Kind = strings.TrimSpace(in.Kind)
	if in.Name == "" {
		return model.Environment{}, invalid("name", "is required")
	}
	if in.Kind == "" {
		in.Kind = "hazard"
	}
	if in.CampaignID <= 0 {
		return model.Environment{}, invalid("campaign_id", "is required")
	}

	id, err := s.repo.CreateEnvironment(ctx, in)
	if err != nil {
		return model.Environment{}, err
	}
	return s.repo.GetEnvironment(ctx, id)
}

func (s *CompendiumService) UpdateEnvironment(ctx context.Context, user model.User, id int64, in model.EnvironmentInput) (model.Environment, error) {
	if !user.IsDM() {
		return model.Environment{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return model.Environment{}, invalid("name", "is required")
	}

	if err := s.repo.UpdateEnvironment(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Environment{}, ErrNotFound
		}
		return model.Environment{}, err
	}
	return s.repo.GetEnvironment(ctx, id)
}

func (s *CompendiumService) DeleteEnvironment(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteEnvironment(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Compendium Items

func (s *CompendiumService) ListItems(ctx context.Context, campaignID int64, category string, q model.Query) (model.Page[model.CompendiumItem], error) {
	items, total, err := s.repo.ListItems(ctx, campaignID, category, q)
	if err != nil {
		return model.Page[model.CompendiumItem]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *CompendiumService) GetItem(ctx context.Context, id int64) (model.CompendiumItem, error) {
	item, err := s.repo.GetItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CompendiumItem{}, ErrNotFound
	}
	return item, err
}

func (s *CompendiumService) CreateItem(ctx context.Context, user model.User, in model.CompendiumItemInput) (model.CompendiumItem, error) {
	if !user.IsDM() {
		return model.CompendiumItem{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Category = strings.TrimSpace(in.Category)
	in.Rarity = strings.TrimSpace(in.Rarity)
	if in.Name == "" {
		return model.CompendiumItem{}, invalid("name", "is required")
	}
	if in.Category == "" {
		in.Category = "gear"
	}
	if in.Rarity == "" {
		in.Rarity = "common"
	}

	id, err := s.repo.CreateItem(ctx, in)
	if err != nil {
		return model.CompendiumItem{}, err
	}
	return s.repo.GetItem(ctx, id)
}

func (s *CompendiumService) UpdateItem(ctx context.Context, user model.User, id int64, in model.CompendiumItemInput) (model.CompendiumItem, error) {
	if !user.IsDM() {
		return model.CompendiumItem{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return model.CompendiumItem{}, invalid("name", "is required")
	}

	if err := s.repo.UpdateItem(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CompendiumItem{}, ErrNotFound
		}
		return model.CompendiumItem{}, err
	}
	return s.repo.GetItem(ctx, id)
}

func (s *CompendiumService) DeleteItem(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Monster Drops

func (s *CompendiumService) ListMonsterDrops(ctx context.Context, monsterID int64) ([]model.MonsterDrop, error) {
	return s.repo.ListMonsterDrops(ctx, monsterID)
}

func (s *CompendiumService) CreateMonsterDrop(ctx context.Context, user model.User, in model.MonsterDropInput) (model.MonsterDrop, error) {
	if !user.IsDM() {
		return model.MonsterDrop{}, ErrForbidden
	}
	in.ItemName = strings.TrimSpace(in.ItemName)
	if in.ItemName == "" {
		return model.MonsterDrop{}, invalid("item_name", "is required")
	}
	if in.DropRate < 1 {
		in.DropRate = 1
	}
	if in.DropRate > 100 {
		in.DropRate = 100
	}
	if in.MinQuantity < 1 {
		in.MinQuantity = 1
	}
	if in.MaxQuantity < in.MinQuantity {
		in.MaxQuantity = in.MinQuantity
	}

	id, err := s.repo.CreateMonsterDrop(ctx, in)
	if err != nil {
		return model.MonsterDrop{}, err
	}
	drops, err := s.repo.ListMonsterDrops(ctx, in.MonsterID)
	if err != nil {
		return model.MonsterDrop{}, err
	}
	for _, d := range drops {
		if d.ID == id {
			return d, nil
		}
	}
	return model.MonsterDrop{}, ErrNotFound
}

func (s *CompendiumService) DeleteMonsterDrop(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteMonsterDrop(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *CompendiumService) RollMonsterLoot(ctx context.Context, monsterID int64) (model.LootRollResult, error) {
	monster, err := s.monsters.Get(ctx, monsterID)
	if err != nil {
		return model.LootRollResult{}, err
	}

	drops, err := s.repo.ListMonsterDrops(ctx, monsterID)
	if err != nil {
		return model.LootRollResult{}, err
	}

	result := model.LootRollResult{
		MonsterName: monster.Name,
		Items:       []model.DroppedItem{},
	}

	for _, drop := range drops {
		roll, _ := rand.Int(rand.Reader, big.NewInt(100))
		d100 := int(roll.Int64()) + 1

		if d100 <= drop.DropRate {
			qty := drop.MinQuantity
			if drop.MaxQuantity > drop.MinQuantity {
				diff := drop.MaxQuantity - drop.MinQuantity + 1
				rQty, _ := rand.Int(rand.Reader, big.NewInt(int64(diff)))
				qty += int(rQty.Int64())
			}
			result.Items = append(result.Items, model.DroppedItem{
				Name:     drop.ItemName,
				Quantity: qty,
			})
			result.CP += drop.CP
			result.SP += drop.SP
			result.GP += drop.GP
		}
	}

	return result, nil
}

// Compendium Spells

func (s *CompendiumService) ListSpells(ctx context.Context, campaignID int64, level int, q model.Query) (model.Page[model.CompendiumSpell], error) {
	items, total, err := s.repo.ListSpells(ctx, campaignID, level, q)
	if err != nil {
		return model.Page[model.CompendiumSpell]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *CompendiumService) GetSpell(ctx context.Context, id int64) (model.CompendiumSpell, error) {
	spell, err := s.repo.GetSpell(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CompendiumSpell{}, ErrNotFound
	}
	return spell, err
}

func (s *CompendiumService) CreateSpell(ctx context.Context, user model.User, in model.CompendiumSpellInput) (model.CompendiumSpell, error) {
	if !user.IsDM() {
		return model.CompendiumSpell{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return model.CompendiumSpell{}, invalid("name", "is required")
	}
	if in.Level < 0 || in.Level > 9 {
		return model.CompendiumSpell{}, invalid("level", "must be between 0 and 9")
	}

	id, err := s.repo.CreateSpell(ctx, in)
	if err != nil {
		return model.CompendiumSpell{}, err
	}
	return s.repo.GetSpell(ctx, id)
}

func (s *CompendiumService) UpdateSpell(ctx context.Context, user model.User, id int64, in model.CompendiumSpellInput) (model.CompendiumSpell, error) {
	if !user.IsDM() {
		return model.CompendiumSpell{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return model.CompendiumSpell{}, invalid("name", "is required")
	}

	if err := s.repo.UpdateSpell(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CompendiumSpell{}, ErrNotFound
		}
		return model.CompendiumSpell{}, err
	}
	return s.repo.GetSpell(ctx, id)
}

func (s *CompendiumService) DeleteSpell(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteSpell(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Status Effects

func (s *CompendiumService) ListStatusEffects(ctx context.Context, campaignID int64, kind string, q model.Query) (model.Page[model.StatusEffect], error) {
	items, total, err := s.repo.ListStatusEffects(ctx, campaignID, kind, q)
	if err != nil {
		return model.Page[model.StatusEffect]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *CompendiumService) GetStatusEffect(ctx context.Context, id int64) (model.StatusEffect, error) {
	effect, err := s.repo.GetStatusEffect(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.StatusEffect{}, ErrNotFound
	}
	return effect, err
}

func (s *CompendiumService) CreateStatusEffect(ctx context.Context, user model.User, in model.StatusEffectInput) (model.StatusEffect, error) {
	if !user.IsDM() {
		return model.StatusEffect{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Kind = strings.TrimSpace(in.Kind)
	if in.Name == "" {
		return model.StatusEffect{}, invalid("name", "is required")
	}
	if in.Kind == "" {
		in.Kind = "condition"
	}
	if in.Color == "" {
		in.Color = "#e67e22"
	}

	id, err := s.repo.CreateStatusEffect(ctx, in)
	if err != nil {
		return model.StatusEffect{}, err
	}
	return s.repo.GetStatusEffect(ctx, id)
}

func (s *CompendiumService) UpdateStatusEffect(ctx context.Context, user model.User, id int64, in model.StatusEffectInput) (model.StatusEffect, error) {
	if !user.IsDM() {
		return model.StatusEffect{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" {
		return model.StatusEffect{}, invalid("name", "is required")
	}

	if err := s.repo.UpdateStatusEffect(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.StatusEffect{}, ErrNotFound
		}
		return model.StatusEffect{}, err
	}
	return s.repo.GetStatusEffect(ctx, id)
}

func (s *CompendiumService) DeleteStatusEffect(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteStatusEffect(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}
