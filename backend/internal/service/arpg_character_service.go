package service

import (
	"context"
	"fmt"
	"strings"

	"dndproject/backend/internal/model"
)

// TransferItem sends an item from one character's bag to another in the same campaign.
func (s *CharacterService) TransferItem(ctx context.Context, user model.User, fromCharID, toCharID, itemID int64, quantity int) error {
	fromChar, err := s.authorize(ctx, user, fromCharID)
	if err != nil {
		return err
	}
	toChar, err := s.repo.Get(ctx, toCharID)
	if err != nil {
		return err
	}
	if fromChar.CampaignID != toChar.CampaignID {
		return invalid("to_character_id", "cannot transfer item across different campaigns")
	}

	item, err := s.findItem(ctx, fromCharID, itemID)
	if err != nil {
		return err
	}
	if quantity <= 0 || quantity > item.Quantity {
		quantity = item.Quantity
	}

	// 1. Reduce or remove from sender
	if quantity >= item.Quantity {
		if err := s.repo.DeleteItem(ctx, itemID); err != nil {
			return err
		}
	} else {
		itemInput := model.CharacterItemInput{
			Name:           item.Name,
			Quantity:       item.Quantity - quantity,
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
		if err := s.repo.UpdateItem(ctx, itemID, itemInput); err != nil {
			return err
		}
	}

	// 2. Add to recipient
	targetInput := model.CharacterItemInput{
		Name:           item.Name,
		Quantity:       quantity,
		Weight:         item.Weight,
		CostCP:         item.CostCP,
		Equipped:       false,
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
	_, err = s.repo.CreateItem(ctx, toCharID, targetInput)
	return err
}

// CheckEquipRequirements verifies that character meets level, class and stat minimums.
func (s *CharacterService) CheckEquipRequirements(char model.Character, item model.CharacterItem) error {
	if char.Level < item.ReqLevel {
		return fmt.Errorf("requires level %d (current: %d)", item.ReqLevel, char.Level)
	}
	if item.ReqClass != "" && !strings.EqualFold(char.Class, item.ReqClass) {
		return fmt.Errorf("requires class %s (current: %s)", item.ReqClass, char.Class)
	}
	if item.ReqStr > 0 && char.Abilities.Strength < item.ReqStr {
		return fmt.Errorf("requires STR %d (current: %d)", item.ReqStr, char.Abilities.Strength)
	}
	if item.ReqDex > 0 && char.Abilities.Dexterity < item.ReqDex {
		return fmt.Errorf("requires DEX %d (current: %d)", item.ReqDex, char.Abilities.Dexterity)
	}
	if item.ReqCon > 0 && char.Abilities.Constitution < item.ReqCon {
		return fmt.Errorf("requires CON %d (current: %d)", item.ReqCon, char.Abilities.Constitution)
	}
	if item.ReqInt > 0 && char.Abilities.Intelligence < item.ReqInt {
		return fmt.Errorf("requires INT %d (current: %d)", item.ReqInt, char.Abilities.Intelligence)
	}
	if item.ReqWis > 0 && char.Abilities.Wisdom < item.ReqWis {
		return fmt.Errorf("requires WIS %d (current: %d)", item.ReqWis, char.Abilities.Wisdom)
	}
	if item.ReqCha > 0 && char.Abilities.Charisma < item.ReqCha {
		return fmt.Errorf("requires CHA %d (current: %d)", item.ReqCha, char.Abilities.Charisma)
	}
	return nil
}

// AllocateStat spends 1 unspent stat point on the specified ability.
func (s *CharacterService) AllocateStat(ctx context.Context, user model.User, charID int64, ability string) (model.CharacterSheet, error) {
	char, err := s.authorize(ctx, user, charID)
	if err != nil {
		return model.CharacterSheet{}, err
	}
	if char.StatPoints <= 0 {
		return model.CharacterSheet{}, invalid("stat_points", "no unspent stat points available")
	}

	char.StatPoints--
	switch strings.ToLower(ability) {
	case "strength", "str":
		char.Abilities.Strength++
	case "dexterity", "dex":
		char.Abilities.Dexterity++
	case "constitution", "con":
		char.Abilities.Constitution++
		// CON increase adds HP retrospectively (standard D&D/ARPG formula)
		char.HPMax += 1
		char.HPCurrent += 1
	case "intelligence", "int":
		char.Abilities.Intelligence++
	case "wisdom", "wis":
		char.Abilities.Wisdom++
	case "charisma", "cha":
		char.Abilities.Charisma++
	default:
		return model.CharacterSheet{}, invalid("ability", "unknown ability score")
	}

	in := characterToInput(char)
	if err := s.repo.Update(ctx, charID, in); err != nil {
		return model.CharacterSheet{}, err
	}
	return s.Sheet(ctx, user, charID)
}

// AddXP awards experience to a character. If eligible for level up, checks DM block.
func (s *CharacterService) AddXP(ctx context.Context, charID int64, amount int64) (model.Character, error) {
	char, err := s.repo.Get(ctx, charID)
	if err != nil {
		return model.Character{}, err
	}
	char.XP += amount
	if char.XP < 0 {
		char.XP = 0
	}

	targetLevel := LevelFromXP(char.XP)
	// Level up progression logic
	if targetLevel > char.Level && !char.LevelBlocked {
		levelsGained := targetLevel - char.Level
		char.Level = targetLevel
		char.StatPoints += levelsGained * 3  // +3 Stat points per level
		char.SkillPoints += levelsGained * 1 // +1 Skill point per level
		// Boost HP based on Hit Dice
		hpGainPerLevel := 5 // Average HP gain per level
		char.HPMax += levelsGained * hpGainPerLevel
		char.HPCurrent += levelsGained * hpGainPerLevel
		char.HitDiceTotal = char.Level
		char.ProficiencyBonus = 2 + (char.Level-1)/4
	}

	in := characterToInput(char)
	if err := s.repo.Update(ctx, charID, in); err != nil {
		return model.Character{}, err
	}
	return s.repo.Get(ctx, charID)
}

// SetLevelBlock allows DM to halt level up progression or release it with an optional reason.
func (s *CharacterService) SetLevelBlock(ctx context.Context, charID int64, blocked bool, reason string) (model.Character, error) {
	char, err := s.repo.Get(ctx, charID)
	if err != nil {
		return model.Character{}, err
	}
	char.LevelBlocked = blocked
	char.BlockReason = reason

	// If unblocking and XP qualifies for level up, trigger level up now
	if !blocked {
		targetLevel := LevelFromXP(char.XP)
		if targetLevel > char.Level {
			levelsGained := targetLevel - char.Level
			char.Level = targetLevel
			char.StatPoints += levelsGained * 3
			char.SkillPoints += levelsGained * 1
			char.HPMax += levelsGained * 5
			char.HPCurrent += levelsGained * 5
			char.HitDiceTotal = char.Level
			char.ProficiencyBonus = 2 + (char.Level-1)/4
		}
	}

	in := characterToInput(char)
	if err := s.repo.Update(ctx, charID, in); err != nil {
		return model.Character{}, err
	}
	return s.repo.Get(ctx, charID)
}

// GrantPoints lets DM grant extra stat or skill points directly.
func (s *CharacterService) GrantPoints(ctx context.Context, charID int64, statPoints, skillPoints int) (model.Character, error) {
	char, err := s.repo.Get(ctx, charID)
	if err != nil {
		return model.Character{}, err
	}
	char.StatPoints += statPoints
	char.SkillPoints += skillPoints
	if char.StatPoints < 0 {
		char.StatPoints = 0
	}
	if char.SkillPoints < 0 {
		char.SkillPoints = 0
	}

	in := characterToInput(char)
	if err := s.repo.Update(ctx, charID, in); err != nil {
		return model.Character{}, err
	}
	return s.repo.Get(ctx, charID)
}
