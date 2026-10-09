package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

const maxSpawnCount = 20

type MonsterService struct {
	repo   *repository.MonsterRepository
	tokens *repository.TokenRepository
}

func NewMonsterService(repo *repository.MonsterRepository, tokens *repository.TokenRepository) *MonsterService {
	return &MonsterService{repo: repo, tokens: tokens}
}

func (s *MonsterService) List(ctx context.Context, q model.Query) (model.Page[model.Monster], error) {
	items, total, err := s.repo.List(ctx, q)
	if err != nil {
		return model.Page[model.Monster]{}, err
	}
	return model.NewPage(items, total, q), nil
}

// StatBlock is the whole entry: vitals, skills and saves, and actions.
func (s *MonsterService) StatBlock(ctx context.Context, id int64) (model.MonsterStatBlock, error) {
	monster, err := s.Get(ctx, id)
	if err != nil {
		return model.MonsterStatBlock{}, err
	}

	proficiencies, err := s.repo.Proficiencies(ctx, id)
	if err != nil {
		return model.MonsterStatBlock{}, err
	}
	actions, err := s.repo.Actions(ctx, id)
	if err != nil {
		return model.MonsterStatBlock{}, err
	}
	return model.MonsterStatBlock{Monster: monster, Proficiencies: proficiencies, Actions: actions}, nil
}

func (s *MonsterService) Get(ctx context.Context, id int64) (model.Monster, error) {
	monster, err := s.repo.Get(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return monster, ErrNotFound
	}
	return monster, err
}

func (s *MonsterService) Create(ctx context.Context, in model.MonsterInput) (model.Monster, error) {
	in, err := normalizeMonster(in)
	if err != nil {
		return model.Monster{}, err
	}
	id, err := s.repo.Create(ctx, in)
	if err != nil {
		return model.Monster{}, err
	}
	return s.Get(ctx, id)
}

func (s *MonsterService) Update(ctx context.Context, id int64, in model.MonsterInput) (model.Monster, error) {
	in, err := normalizeMonster(in)
	if err != nil {
		return model.Monster{}, err
	}
	if err := s.repo.Update(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Monster{}, ErrNotFound
		}
		return model.Monster{}, err
	}
	return s.Get(ctx, id)
}

func (s *MonsterService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *MonsterService) SetProficiency(ctx context.Context, monsterID int64, p model.MonsterProficiency, keep bool) ([]model.MonsterProficiency, error) {
	if _, err := s.Get(ctx, monsterID); err != nil {
		return nil, err
	}
	p.Name = strings.ToLower(strings.TrimSpace(p.Name))
	if err := validProficiency(p.Kind, p.Name); err != nil {
		return nil, err
	}
	if err := s.repo.SetProficiency(ctx, monsterID, p, keep); err != nil {
		return nil, err
	}
	return s.repo.Proficiencies(ctx, monsterID)
}

func (s *MonsterService) AddAction(ctx context.Context, monsterID int64, in model.MonsterActionInput) (model.MonsterAction, error) {
	if _, err := s.Get(ctx, monsterID); err != nil {
		return model.MonsterAction{}, err
	}
	in, err := normalizeAction(in)
	if err != nil {
		return model.MonsterAction{}, err
	}

	id, err := s.repo.CreateAction(ctx, monsterID, in)
	if err != nil {
		return model.MonsterAction{}, err
	}
	return s.repo.Action(ctx, id)
}

func (s *MonsterService) UpdateAction(ctx context.Context, id int64, in model.MonsterActionInput) (model.MonsterAction, error) {
	in, err := normalizeAction(in)
	if err != nil {
		return model.MonsterAction{}, err
	}
	if err := s.repo.UpdateAction(ctx, id, in); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.MonsterAction{}, ErrNotFound
		}
		return model.MonsterAction{}, err
	}
	return s.repo.Action(ctx, id)
}

func (s *MonsterService) DeleteAction(ctx context.Context, id int64) error {
	err := s.repo.DeleteAction(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

// Spawn places copies of a monster on a map. Vitals are copied onto each token
// rather than read through the link, so two goblins take damage separately
// while still sharing one stat block for their actions.
func (s *MonsterService) Spawn(ctx context.Context, in model.SpawnInput) ([]model.Token, error) {
	monster, err := s.Get(ctx, in.MonsterID)
	if err != nil {
		return nil, err
	}
	if in.MapID <= 0 {
		return nil, invalid("map_id", "is required")
	}

	count := in.Count
	if count < 1 {
		count = 1
	}
	if count > maxSpawnCount {
		return nil, invalid("count", fmt.Sprintf("must be at most %d", maxSpawnCount))
	}

	spawned := make([]model.Token, 0, count)
	for i := range count {
		hp := monster.HPAverage
		if in.RollHP && monster.HPFormula != "" {
			result, err := EvaluateRoll(monster.HPFormula)
			if err != nil {
				return nil, invalid("hp_formula", "the stat block hit point formula is not valid")
			}
			hp = max(1, result.Total)
		}

		name := monster.Name
		if count > 1 {
			// Numbered so the initiative order can tell them apart.
			name = fmt.Sprintf("%s %d", monster.Name, i+1)
		}

		monsterID := monster.ID
		id, err := s.tokens.Create(ctx, model.TokenInput{
			MapID:      in.MapID,
			MonsterID:  &monsterID,
			Name:       name,
			ImageURL:   monster.ImageURL,
			Color:      monster.Color,
			Size:       max(1, monster.TokenSize),
			HPCurrent:  hp,
			HPMax:      hp,
			ArmorClass: monster.ArmorClass,
			// Stagger them so a group does not land in one pile.
			X: in.X + float64(i%5)*60,
			Y: in.Y + float64(i/5)*60,
		})
		if err != nil {
			return nil, err
		}

		token, err := s.tokens.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		token.Conditions = []string{}
		spawned = append(spawned, token)
	}
	return spawned, nil
}

func validProficiency(kind, name string) error {
	switch kind {
	case "skill":
		if _, ok := model.Skills[name]; !ok {
			return invalid("name", "is not a recognised skill")
		}
	case "save":
		if _, ok := model.Skills[name]; ok || !isSave(name) {
			return invalid("name", "must be one of the six abilities")
		}
	default:
		return invalid("kind", "must be skill or save")
	}
	return nil
}

func isSave(name string) bool {
	for _, save := range model.SaveNames {
		if save == name {
			return true
		}
	}
	return false
}

func normalizeMonster(in model.MonsterInput) (model.MonsterInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Type = strings.TrimSpace(in.Type)
	in.Traits = strings.TrimSpace(in.Traits)
	in.Color = strings.TrimSpace(in.Color)

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Size == "" {
		in.Size = "medium"
	}
	if in.Color == "" {
		in.Color = "#8e5a3b"
	}
	if in.ArmorClass <= 0 {
		in.ArmorClass = 10
	}
	if in.HPAverage < 1 {
		in.HPAverage = 1
	}
	if in.Speed < 0 {
		in.Speed = 30
	}
	if in.TokenSize < 1 {
		in.TokenSize = 1
	}
	if in.ChallengeRating == "" {
		in.ChallengeRating = "0"
	}

	// The formula is checked here so a bad one fails when the stat block is
	// written, not in the middle of a fight.
	if in.HPFormula != "" {
		if _, err := EvaluateRoll(in.HPFormula); err != nil {
			return in, invalid("hp_formula", "is not a valid dice expression")
		}
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

func normalizeAction(in model.MonsterActionInput) (model.MonsterActionInput, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Damage = strings.TrimSpace(in.Damage)
	in.Description = strings.TrimSpace(in.Description)

	if in.Name == "" {
		return in, invalid("name", "is required")
	}
	if in.Damage != "" {
		if _, err := EvaluateRoll(in.Damage); err != nil {
			return in, invalid("damage", "is not a valid dice expression")
		}
	}
	return in, nil
}
