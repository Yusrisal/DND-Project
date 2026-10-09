package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type ShopService struct {
	repo       *repository.ShopRepository
	characters *repository.CharacterRepository
}

func NewShopService(repo *repository.ShopRepository, characters *repository.CharacterRepository) *ShopService {
	return &ShopService{repo: repo, characters: characters}
}

func (s *ShopService) List(ctx context.Context, campaignID int64, q model.Query) (model.Page[model.Shop], error) {
	items, total, err := s.repo.List(ctx, campaignID, q)
	if err != nil {
		return model.Page[model.Shop]{}, err
	}
	return model.NewPage(items, total, q), nil
}

func (s *ShopService) Get(ctx context.Context, id int64) (model.ShopDetails, error) {
	shop, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ShopDetails{}, ErrNotFound
	}
	if err != nil {
		return model.ShopDetails{}, err
	}
	items, err := s.repo.Items(ctx, id)
	if err != nil {
		return model.ShopDetails{}, err
	}
	return model.ShopDetails{Shop: shop, Items: items}, nil
}

func (s *ShopService) Create(ctx context.Context, user model.User, in model.ShopInput) (model.Shop, error) {
	if !user.IsDM() {
		return model.Shop{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return model.Shop{}, invalid("name", "is required")
	}
	if in.CampaignID <= 0 {
		return model.Shop{}, invalid("campaign_id", "is required")
	}

	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Shop{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *ShopService) Update(ctx context.Context, user model.User, id int64, in model.ShopInput) (model.Shop, error) {
	if !user.IsDM() {
		return model.Shop{}, ErrForbidden
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return model.Shop{}, invalid("name", "is required")
	}

	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Shop{}, ErrNotFound
		}
		return model.Shop{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *ShopService) Delete(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *ShopService) AddItem(ctx context.Context, user model.User, shopID int64, in model.ShopItemInput) (model.ShopItem, error) {
	if !user.IsDM() {
		return model.ShopItem{}, ErrForbidden
	}
	in, err := normalizeShopItem(in)
	if err != nil {
		return model.ShopItem{}, err
	}

	id, err := s.repo.CreateItem(ctx, shopID, in)
	if err != nil {
		return model.ShopItem{}, err
	}
	return s.repo.Item(ctx, id)
}

func (s *ShopService) UpdateItem(ctx context.Context, user model.User, id int64, in model.ShopItemInput) (model.ShopItem, error) {
	if !user.IsDM() {
		return model.ShopItem{}, ErrForbidden
	}
	in, err := normalizeShopItem(in)
	if err != nil {
		return model.ShopItem{}, err
	}

	if err := s.repo.UpdateItem(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.ShopItem{}, ErrNotFound
		}
		return model.ShopItem{}, err
	}
	return s.repo.Item(ctx, id)
}

func (s *ShopService) DeleteItem(ctx context.Context, user model.User, id int64) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	err := s.repo.DeleteItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *ShopService) BuyItem(ctx context.Context, user model.User, shopItemID int64, req model.BuyItemRequest) (model.CharacterItem, error) {
	if req.CharacterID <= 0 {
		return model.CharacterItem{}, invalid("character_id", "is required")
	}
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	// Verify caller owns character or is DM
	char, err := s.characters.Get(ctx, req.CharacterID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.CharacterItem{}, ErrNotFound
	}
	if err != nil {
		return model.CharacterItem{}, err
	}
	if !user.IsDM() && (char.UserID == nil || *char.UserID != user.ID) {
		return model.CharacterItem{}, ErrForbidden
	}

	createdID, err := s.repo.BuyItem(ctx, shopItemID, req.CharacterID, req.Quantity)
	if err != nil {
		return model.CharacterItem{}, invalid("purchase", err.Error())
	}

	return s.characters.Item(ctx, createdID)
}

func normalizeShopItem(in model.ShopItemInput) (model.ShopItemInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Damage = strings.TrimSpace(in.Damage)
	in.DamageType = strings.TrimSpace(in.DamageType)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.CostCP < 0 {
		in.CostCP = 0
	}
	if in.Weight < 0 {
		return in, invalid("weight", "cannot be negative")
	}
	return in, nil
}
