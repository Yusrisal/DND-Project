package service

import (
	"context"
	"fmt"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

type SkillService struct {
	repo       *repository.SkillRepository
	characters *repository.CharacterRepository
}

func NewSkillService(repo *repository.SkillRepository, characters *repository.CharacterRepository) *SkillService {
	return &SkillService{repo: repo, characters: characters}
}

func (s *SkillService) ListClassSkills(ctx context.Context, className string) ([]model.ClassSkill, error) {
	return s.repo.ListClassSkills(ctx, className)
}

func (s *SkillService) CharacterSkills(ctx context.Context, charID int64) ([]model.CharacterLearnedSkill, error) {
	return s.repo.CharacterSkills(ctx, charID)
}

// LearnSkill unlocks a class skill node using 1 skill point, enforcing level and tree prerequisites.
func (s *SkillService) LearnSkill(ctx context.Context, user model.User, charID, skillID int64) error {
	char, err := s.characters.Get(ctx, charID)
	if err != nil {
		return err
	}
	if !user.IsDM() && (char.UserID == nil || *char.UserID != user.ID) {
		return ErrForbidden
	}

	skill, err := s.repo.GetClassSkill(ctx, skillID)
	if err != nil {
		return err
	}

	// Class verification
	if !strings.EqualFold(char.Class, skill.ClassName) {
		return invalid("class", fmt.Sprintf("skill belongs to class %s (character is %s)", skill.ClassName, char.Class))
	}

	// Level requirement
	if char.Level < skill.ReqLevel {
		return invalid("req_level", fmt.Sprintf("requires level %d (current: %d)", skill.ReqLevel, char.Level))
	}

	// Check if already learned and can level up
	learned, err := s.repo.CharacterSkills(ctx, charID)
	if err != nil {
		return err
	}
	var existing *model.CharacterLearnedSkill
	for _, l := range learned {
		if l.SkillID == skillID {
			existing = &l
			break
		}
	}

	if existing != nil {
		if existing.CurrentRank >= skill.MaxRank {
			return invalid("rank", fmt.Sprintf("skill is already at maximum rank (%d)", skill.MaxRank))
		}
	} else {
		// New skill prerequisites check
		if skill.ParentSkillID != nil {
			parentLearned := false
			for _, l := range learned {
				if l.SkillID == *skill.ParentSkillID {
					parentLearned = true
					break
				}
			}
			if !parentLearned {
				return invalid("parent_skill", "prerequisite skill not unlocked yet")
			}
		}
	}

	// Check and deduct skill points
	if char.SkillPoints < 1 {
		return invalid("skill_points", "insufficient skill points")
	}

	char.SkillPoints--
	in := characterToInput(char)
	if err := s.characters.Update(ctx, charID, in); err != nil {
		return err
	}

	return s.repo.LearnSkill(ctx, charID, skillID)
}

// ToggleSkillActive equips/unequips a learned skill into the active combat deck (max 10).
func (s *SkillService) ToggleSkillActive(ctx context.Context, user model.User, charID, skillID int64, active bool) error {
	char, err := s.characters.Get(ctx, charID)
	if err != nil {
		return err
	}
	if !user.IsDM() && (char.UserID == nil || *char.UserID != user.ID) {
		return ErrForbidden
	}

	return s.repo.SetSkillActive(ctx, charID, skillID, active)
}

// UseSkillScroll consumes a learning scroll item to teach a skill directly.
func (s *SkillService) UseSkillScroll(ctx context.Context, user model.User, charID, itemID int64) error {
	char, err := s.characters.Get(ctx, charID)
	if err != nil {
		return err
	}
	if !user.IsDM() && (char.UserID == nil || *char.UserID != user.ID) {
		return ErrForbidden
	}

	item, err := s.characters.Item(ctx, itemID)
	if err != nil {
		return err
	}
	if item.TeachesSkillID == nil {
		return invalid("item", "this item is not a learning scroll")
	}

	// Learn skill without charging skill point
	if err := s.repo.LearnSkill(ctx, charID, *item.TeachesSkillID); err != nil {
		return err
	}

	// Consume scroll
	if item.Quantity <= 1 {
		return s.characters.DeleteItem(ctx, itemID)
	}
	itemInput := model.CharacterItemInput{
		Name:           item.Name,
		Quantity:       item.Quantity - 1,
		Weight:         item.Weight,
		CostCP:         item.CostCP,
		Equipped:       item.Equipped,
		AttackBonus:    item.AttackBonus,
		Damage:         item.Damage,
		DamageType:     item.DamageType,
		Description:    item.Description,
		ReqLevel:       item.ReqLevel,
		ReqClass:       item.ReqClass,
		ReqStr:         item.ReqStr,
		ReqDex:         item.ReqDex,
		ReqCon:         item.ReqCon,
		ReqInt:         item.ReqInt,
		ReqWis:         item.ReqWis,
		ReqCha:         item.ReqCha,
		TeachesSkillID: item.TeachesSkillID,
	}
	return s.characters.UpdateItem(ctx, itemID, itemInput)
}

// Pending Loot Drops (DM Queue)
func (s *SkillService) ListPendingLoot(ctx context.Context, campaignID int64) ([]model.PendingLootDrop, error) {
	return s.repo.ListPendingLoot(ctx, campaignID)
}

func (s *SkillService) ResolvePendingLoot(ctx context.Context, dropID int64, approve bool, reason string) error {
	drop, err := s.repo.GetPendingLoot(ctx, dropID)
	if err != nil {
		return err
	}

	if approve && drop.AssignedToCharacterID != nil {
		// Award item to character bag
		itemInput := model.CharacterItemInput{
			Name:        drop.ItemName,
			Quantity:    drop.Quantity,
			Weight:      1.0,
			CostCP:      drop.CP + drop.SP*10 + drop.GP*100,
			Equipped:    false,
			Description: fmt.Sprintf("Dropped by %s", drop.MonsterName),
			ReqLevel:    1,
		}
		if _, err := s.characters.CreateItem(ctx, *drop.AssignedToCharacterID, itemInput); err != nil {
			return err
		}
	}

	return s.repo.ResolvePendingLoot(ctx, dropID, approve, reason)
}
