package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// 5e bounds. They are generous rather than exact: the point is to stop a typo
// or a crafted request writing nonsense, not to referee the rules.
const (
	maxCharacterLevel = 50
	maxAbilityScore   = 50
	maxSpellLevel     = 9
	maxSlotCount      = 20
)

type CharacterService struct {
	repo   *repository.CharacterRepository
	skills *repository.SkillRepository
}

func NewCharacterService(repo *repository.CharacterRepository, skills *repository.SkillRepository) *CharacterService {
	return &CharacterService{repo: repo, skills: skills}
}

// List shows the DM every sheet in the campaign and a player only their own.
func (s *CharacterService) List(ctx context.Context, user model.User, campaignID int64, q model.Query) (model.Page[model.Character], error) {
	ownerFilter := int64(0)
	if !user.IsDM() {
		ownerFilter = user.ID
	}

	items, total, err := s.repo.List(ctx, campaignID, ownerFilter, q)
	if err != nil {
		return model.Page[model.Character]{}, err
	}
	return model.NewPage(items, total, q), nil
}

// Sheet returns the character with its items, spells and slots in one response.
func (s *CharacterService) Sheet(ctx context.Context, user model.User, id int64) (model.CharacterSheet, error) {
	character, err := s.authorize(ctx, user, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}

	items, err := s.repo.Items(ctx, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}
	spells, err := s.repo.Spells(ctx, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}
	var skillsLoaded []model.CharacterLearnedSkill
	if s.skills != nil {
		skillsLoaded, _ = s.skills.CharacterSkills(ctx, id)
	}
	slots, err := s.repo.Slots(ctx, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}
	proficiencies, err := s.repo.Proficiencies(ctx, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}

	skills, saves := model.Checks(character, proficiencies)
	return model.CharacterSheet{
		Character:     character,
		Items:         items,
		Spells:        spells,
		SkillsLoaded:  skillsLoaded,
		Slots:         slots,
		Proficiencies: proficiencies,
		Skills:        skills,
		Saves:         saves,
	}, nil
}

// SetProficiency marks a skill or save as trained, or clears it. Expertise
// doubles the proficiency bonus, which is how rogues and bards work.
func (s *CharacterService) SetProficiency(ctx context.Context, user model.User, characterID int64, p model.Proficiency, trained bool) (model.CharacterSheet, error) {
	if _, err := s.authorize(ctx, user, characterID); err != nil {
		return model.CharacterSheet{}, err
	}
	p.Name = strings.ToLower(strings.TrimSpace(p.Name))
	if err := validProficiency(p.Kind, p.Name); err != nil {
		return model.CharacterSheet{}, err
	}
	if err := s.repo.SetProficiency(ctx, characterID, p, trained); err != nil {
		return model.CharacterSheet{}, err
	}
	return s.Sheet(ctx, user, characterID)
}

func (s *CharacterService) Get(ctx context.Context, user model.User, id int64) (model.Character, error) {
	return s.authorize(ctx, user, id)
}

// Create assigns the sheet to the caller unless the DM says otherwise, so a
// player cannot make a character for someone else.
func (s *CharacterService) Create(ctx context.Context, user model.User, in model.CharacterInput) (model.Character, error) {
	in, err := normalizeCharacter(in)
	if err != nil {
		return model.Character{}, err
	}
	if in.CampaignID <= 0 {
		return model.Character{}, invalid("campaign_id", "is required")
	}
	if !user.IsDM() {
		owner := user.ID
		in.UserID = &owner
	}

	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Character{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *CharacterService) Update(ctx context.Context, user model.User, id int64, in model.CharacterInput) (model.Character, error) {
	current, err := s.authorize(ctx, user, id)
	if err != nil {
		return model.Character{}, err
	}

	in, err = normalizeCharacter(in)
	if err != nil {
		return model.Character{}, err
	}
	// Reassigning a sheet is a DM move; a player editing their own keeps the owner.
	if !user.IsDM() {
		in.UserID = current.UserID
	}

	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Character{}, ErrNotFound
		}
		return model.Character{}, err
	}
	return s.repo.Get(ctx, id)
}

func (s *CharacterService) Delete(ctx context.Context, user model.User, id int64) error {
	if _, err := s.authorize(ctx, user, id); err != nil {
		return err
	}
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Rest & Death Saves

func (s *CharacterService) LongRest(ctx context.Context, user model.User, id int64) (model.CharacterSheet, error) {
	if _, err := s.authorize(ctx, user, id); err != nil {
		return model.CharacterSheet{}, err
	}
	if err := s.repo.LongRest(ctx, id); err != nil {
		return model.CharacterSheet{}, err
	}
	return s.Sheet(ctx, user, id)
}

func (s *CharacterService) ShortRest(ctx context.Context, user model.User, id int64, hitDiceSpent, hpHealed int) (model.CharacterSheet, error) {
	char, err := s.authorize(ctx, user, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}
	if hitDiceSpent < 0 || hitDiceSpent+char.HitDiceUsed > char.HitDiceTotal {
		return model.CharacterSheet{}, invalid("hit_dice_spent", "cannot spend more hit dice than available")
	}
	if hpHealed < 0 {
		hpHealed = 0
	}
	if err := s.repo.ShortRest(ctx, id, hitDiceSpent, hpHealed); err != nil {
		return model.CharacterSheet{}, err
	}
	return s.Sheet(ctx, user, id)
}

func (s *CharacterService) DeathSave(ctx context.Context, user model.User, id int64, roll int) (model.CharacterSheet, error) {
	char, err := s.authorize(ctx, user, id)
	if err != nil {
		return model.CharacterSheet{}, err
	}

	if roll == 20 {
		// Natural 20: creature regains 1 HP
		char.HPCurrent = 1
		char.DeathSaveSuccesses = 0
		char.DeathSaveFailures = 0
	} else if roll >= 10 {
		char.DeathSaveSuccesses++
	} else if roll == 1 {
		// Natural 1: two failures
		char.DeathSaveFailures += 2
	} else {
		char.DeathSaveFailures++
	}

	if char.DeathSaveFailures >= 3 {
		char.Status = "dead"
	}

	in := characterToInput(char)
	if err := s.repo.Update(ctx, id, in); err != nil {
		return model.CharacterSheet{}, err
	}
	return s.Sheet(ctx, user, id)
}

func (s *CharacterService) SplitPartyLoot(ctx context.Context, user model.User, req model.LootSplitRequest) error {
	if !user.IsDM() {
		return ErrForbidden
	}
	if req.CampaignID <= 0 {
		return invalid("campaign_id", "is required")
	}
	return s.repo.SplitPartyLoot(ctx, req.CampaignID, req.CP, req.SP, req.EP, req.GP, req.PP)
}

// Items

func (s *CharacterService) AddItem(ctx context.Context, user model.User, characterID int64, in model.CharacterItemInput) (model.CharacterItem, error) {
	if _, err := s.authorize(ctx, user, characterID); err != nil {
		return model.CharacterItem{}, err
	}
	in, err := normalizeItem(in)
	if err != nil {
		return model.CharacterItem{}, err
	}

	id, err := s.repo.CreateItem(ctx, characterID, in)
	if err != nil {
		return model.CharacterItem{}, err
	}
	return s.findItem(ctx, characterID, id)
}

func (s *CharacterService) UpdateItem(ctx context.Context, user model.User, id int64, in model.CharacterItemInput) (model.CharacterItem, error) {
	characterID, err := s.authorizeItem(ctx, user, id)
	if err != nil {
		return model.CharacterItem{}, err
	}
	in, err = normalizeItem(in)
	if err != nil {
		return model.CharacterItem{}, err
	}

	if err := s.repo.UpdateItem(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CharacterItem{}, ErrNotFound
		}
		return model.CharacterItem{}, err
	}
	return s.findItem(ctx, characterID, id)
}

func (s *CharacterService) DeleteItem(ctx context.Context, user model.User, id int64) error {
	if _, err := s.authorizeItem(ctx, user, id); err != nil {
		return err
	}
	err := s.repo.DeleteItem(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Spells

func (s *CharacterService) AddSpell(ctx context.Context, user model.User, characterID int64, in model.CharacterSpellInput) (model.CharacterSpell, error) {
	if _, err := s.authorize(ctx, user, characterID); err != nil {
		return model.CharacterSpell{}, err
	}
	in, err := normalizeSpell(in)
	if err != nil {
		return model.CharacterSpell{}, err
	}

	id, err := s.repo.CreateSpell(ctx, characterID, in)
	if err != nil {
		return model.CharacterSpell{}, err
	}
	return s.findSpell(ctx, characterID, id)
}

func (s *CharacterService) UpdateSpell(ctx context.Context, user model.User, id int64, in model.CharacterSpellInput) (model.CharacterSpell, error) {
	characterID, err := s.authorizeSpell(ctx, user, id)
	if err != nil {
		return model.CharacterSpell{}, err
	}
	in, err = normalizeSpell(in)
	if err != nil {
		return model.CharacterSpell{}, err
	}

	if err := s.repo.UpdateSpell(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.CharacterSpell{}, ErrNotFound
		}
		return model.CharacterSpell{}, err
	}
	return s.findSpell(ctx, characterID, id)
}

func (s *CharacterService) DeleteSpell(ctx context.Context, user model.User, id int64) error {
	if _, err := s.authorizeSpell(ctx, user, id); err != nil {
		return err
	}
	err := s.repo.DeleteSpell(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// SetSlot writes one spell level. Spending a slot is the same call with used
// raised by one, which keeps the server as the only place the count changes.
func (s *CharacterService) SetSlot(ctx context.Context, user model.User, characterID int64, level int, in model.SpellSlotInput) ([]model.SpellSlot, error) {
	if _, err := s.authorize(ctx, user, characterID); err != nil {
		return nil, err
	}
	if level < 1 || level > maxSpellLevel {
		return nil, invalid("level", "must be between 1 and 9")
	}
	if in.Total < 0 || in.Total > maxSlotCount {
		return nil, invalid("total", "must be between 0 and 20")
	}
	if in.Used < 0 || in.Used > in.Total {
		return nil, invalid("used", "must be between 0 and the total")
	}

	if err := s.repo.SetSlot(ctx, characterID, level, in.Total, in.Used); err != nil {
		return nil, err
	}
	return s.repo.Slots(ctx, characterID)
}

// authorize is the single ownership gate: the DM reaches every sheet, a player
// only their own. Every mutating call above goes through it.
func (s *CharacterService) authorize(ctx context.Context, user model.User, characterID int64) (model.Character, error) {
	character, err := s.repo.Get(ctx, characterID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Character{}, ErrNotFound
	}
	if err != nil {
		return model.Character{}, err
	}
	if user.IsDM() {
		return character, nil
	}
	if character.UserID == nil || *character.UserID != user.ID {
		return model.Character{}, ErrForbidden
	}
	return character, nil
}

func (s *CharacterService) authorizeItem(ctx context.Context, user model.User, itemID int64) (int64, error) {
	characterID, err := s.repo.ItemCharacter(ctx, itemID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	_, err = s.authorize(ctx, user, characterID)
	return characterID, err
}

func (s *CharacterService) authorizeSpell(ctx context.Context, user model.User, spellID int64) (int64, error) {
	characterID, err := s.repo.SpellCharacter(ctx, spellID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	_, err = s.authorize(ctx, user, characterID)
	return characterID, err
}

func (s *CharacterService) findItem(ctx context.Context, characterID, id int64) (model.CharacterItem, error) {
	items, err := s.repo.Items(ctx, characterID)
	if err != nil {
		return model.CharacterItem{}, err
	}
	for _, item := range items {
		if item.ID == id {
			return item, nil
		}
	}
	return model.CharacterItem{}, ErrNotFound
}

func (s *CharacterService) findSpell(ctx context.Context, characterID, id int64) (model.CharacterSpell, error) {
	spells, err := s.repo.Spells(ctx, characterID)
	if err != nil {
		return model.CharacterSpell{}, err
	}
	for _, spell := range spells {
		if spell.ID == id {
			return spell, nil
		}
	}
	return model.CharacterSpell{}, ErrNotFound
}

func normalizeCharacter(in model.CharacterInput) (model.CharacterInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Class = strings.TrimSpace(in.Class)
	in.Race = strings.TrimSpace(in.Race)
	in.Notes = strings.TrimSpace(in.Notes)
	in.Status = strings.ToLower(strings.TrimSpace(in.Status))

	if in.Status == "" {
		in.Status = "active"
	}
	if in.Status != "active" && in.Status != "tavern" && in.Status != "dead" && in.Status != "retired" {
		return in, invalid("status", "must be active, tavern, dead, or retired")
	}

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Level < 1 {
		in.Level = 1
	}
	if in.Level > maxCharacterLevel {
		return in, invalid("level", "must be between 1 and 50")
	}
	if in.HPMax < 0 || in.HPCurrent < 0 || in.HPTemp < 0 {
		return in, invalid("hp_max", "cannot be negative")
	}
	if in.ArmorClass <= 0 {
		in.ArmorClass = 10
	}
	if in.Speed < 0 {
		in.Speed = 30
	}
	if in.ProficiencyBonus <= 0 {
		in.ProficiencyBonus = 2 + (in.Level-1)/4
	}
	if in.HitDiceDie == "" {
		in.HitDiceDie = "d8"
	}
	if in.HitDiceTotal <= 0 {
		in.HitDiceTotal = max(1, in.Level)
	}
	if in.CP < 0 || in.SP < 0 || in.EP < 0 || in.GP < 0 || in.PP < 0 {
		return in, invalid("currency", "cannot have negative coins")
	}

	scores := []int{
		in.Abilities.Strength, in.Abilities.Dexterity, in.Abilities.Constitution,
		in.Abilities.Intelligence, in.Abilities.Wisdom, in.Abilities.Charisma,
	}
	for _, score := range scores {
		if score < 1 || score > maxAbilityScore {
			return in, invalid("abilities", "each score must be between 1 and 30")
		}
	}
	return in, nil
}

func normalizeItem(in model.CharacterItemInput) (model.CharacterItemInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Quantity < 1 {
		in.Quantity = 1
	}
	if in.Weight < 0 {
		return in, invalid("weight", "cannot be negative")
	}
	if in.CostCP < 0 {
		in.CostCP = 0
	}
	return in, nil
}

func normalizeSpell(in model.CharacterSpellInput) (model.CharacterSpellInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.School = strings.TrimSpace(in.School)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Level < 0 || in.Level > maxSpellLevel {
		return in, invalid("level", "must be between 0 and 9")
	}
	return in, nil
}

func characterToInput(c model.Character) model.CharacterInput {
	return model.CharacterInput{
		CampaignID:            c.CampaignID,
		UserID:                c.UserID,
		Status:                c.Status,
		Name:                  c.Name,
		Class:                 c.Class,
		Race:                  c.Race,
		Level:                 c.Level,
		XP:                    c.XP,
		StatPoints:            c.StatPoints,
		SkillPoints:           c.SkillPoints,
		LevelBlocked:          c.LevelBlocked,
		BlockReason:           c.BlockReason,
		HPCurrent:             c.HPCurrent,
		HPMax:                 c.HPMax,
		HPTemp:                c.HPTemp,
		ArmorClass:            c.ArmorClass,
		Speed:                 c.Speed,
		ProficiencyBonus:      c.ProficiencyBonus,
		CP:                    c.CP,
		SP:                    c.SP,
		EP:                    c.EP,
		GP:                    c.GP,
		PP:                    c.PP,
		Abilities:             c.Abilities,
		Notes:                 c.Notes,
		DeathSaveSuccesses:    c.DeathSaveSuccesses,
		DeathSaveFailures:     c.DeathSaveFailures,
		HitDiceTotal:          c.HitDiceTotal,
		HitDiceUsed:           c.HitDiceUsed,
		HitDiceDie:            c.HitDiceDie,
		DamageResistances:     c.DamageResistances,
		DamageVulnerabilities: c.DamageVulnerabilities,
		DamageImmunities:      c.DamageImmunities,
		ConditionImmunities:   c.ConditionImmunities,
	}
}
