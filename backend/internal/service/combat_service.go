package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"dndproject/backend/internal/model"
	"dndproject/backend/internal/repository"
)

// RollKindAttack tags a stored roll whose detail is a full attack resolution,
// so the feed can render it as an attack card and history works for free.
const RollKindAttack = "attack"

type CombatService struct {
	encounters *repository.EncounterRepository
	tokens     *repository.TokenRepository
	characters *repository.CharacterRepository
	monsters   *repository.MonsterRepository
	sheets     *CharacterService
	rolls      *RollService
}

func NewCombatService(
	encounters *repository.EncounterRepository,
	tokens *repository.TokenRepository,
	characters *repository.CharacterRepository,
	monsters *repository.MonsterRepository,
	sheets *CharacterService,
	rolls *RollService,
) *CombatService {
	return &CombatService{
		encounters: encounters, tokens: tokens, characters: characters,
		monsters: monsters, sheets: sheets, rolls: rolls,
	}
}

// attackSource is one resolved attack, whether it came off a character sheet or
// a monster stat block, so the resolution below does not care which.
type attackSource struct {
	attackerName string
	label        string
	attackBonus  int
	damage       string
	damageType   string
	characterID  *int64
}

// Encounter returns the running encounter for a map, or ErrNotFound when combat
// is not underway. A quiet map is the normal case, so callers treat that as
// "no initiative order" rather than as a failure.
func (s *CombatService) Encounter(ctx context.Context, mapID int64) (model.Encounter, error) {
	encounter, err := s.encounters.Active(ctx, mapID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Encounter{}, ErrNotFound
	}
	return encounter, err
}

// Start opens an encounter on a map. Any encounter already running there is
// closed first, so a map never has two initiative orders.
func (s *CombatService) Start(ctx context.Context, mapID int64) (model.Encounter, error) {
	if existing, err := s.encounters.Active(ctx, mapID); err == nil {
		if err := s.encounters.End(ctx, existing.ID); err != nil {
			return model.Encounter{}, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return model.Encounter{}, err
	}

	id, err := s.encounters.Create(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, id)
}

func (s *CombatService) End(ctx context.Context, mapID int64) (model.Encounter, error) {
	encounter, err := s.Encounter(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}
	if err := s.encounters.End(ctx, encounter.ID); err != nil {
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, encounter.ID)
}

// AddCombatant puts a token, or a bare name, into the order. An initiative of
// zero means "roll it": 1d20 plus the dexterity modifier of the linked sheet.
func (s *CombatService) AddCombatant(ctx context.Context, mapID int64, in model.CombatantInput) (model.Encounter, error) {
	encounter, err := s.Encounter(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}

	in.Name = strings.TrimSpace(in.Name)
	if in.TokenID != nil {
		token, err := s.tokens.Get(ctx, *in.TokenID)
		if errors.Is(err, sql.ErrNoRows) {
			return model.Encounter{}, ErrNotFound
		}
		if err != nil {
			return model.Encounter{}, err
		}
		if token.MapID != mapID {
			return model.Encounter{}, invalid("token_id", "is not on this map")
		}
		if in.Name == "" {
			in.Name = token.Name
		}
		if in.Initiative == 0 {
			if in.Initiative, err = s.rollInitiative(ctx, token); err != nil {
				return model.Encounter{}, err
			}
		}
	}

	if in.Name == "" {
		return model.Encounter{}, invalid("name", "is required")
	}
	if err := s.encounters.AddCombatant(ctx, encounter.ID, in); err != nil {
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, encounter.ID)
}

func (s *CombatService) SetInitiative(ctx context.Context, mapID, combatantID int64, initiative int) (model.Encounter, error) {
	encounter, err := s.Encounter(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}
	if err := s.encounters.SetInitiative(ctx, combatantID, initiative); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Encounter{}, ErrNotFound
		}
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, encounter.ID)
}

// RemoveCombatant drops a line from the order and keeps the turn pointer on the
// same creature, so removing a dead monster does not skip anyone.
func (s *CombatService) RemoveCombatant(ctx context.Context, mapID, combatantID int64) (model.Encounter, error) {
	encounter, err := s.Encounter(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}

	removedIndex := -1
	for i, c := range encounter.Combatants {
		if c.ID == combatantID {
			removedIndex = i
			break
		}
	}

	if _, err := s.encounters.RemoveCombatant(ctx, combatantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Encounter{}, ErrNotFound
		}
		return model.Encounter{}, err
	}

	turn := encounter.TurnIndex
	if removedIndex >= 0 && removedIndex < turn {
		turn--
	}
	remaining := len(encounter.Combatants) - 1
	if remaining <= 0 {
		turn = 0
	} else if turn >= remaining {
		turn = 0
	}
	if err := s.encounters.SetTurn(ctx, encounter.ID, encounter.Round, turn); err != nil {
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, encounter.ID)
}

// NextTurn advances the pointer, rolling over into the next round at the end of
// the order and decrementing active condition durations.
func (s *CombatService) NextTurn(ctx context.Context, mapID int64) (model.Encounter, error) {
	encounter, err := s.Encounter(ctx, mapID)
	if err != nil {
		return model.Encounter{}, err
	}
	if len(encounter.Combatants) == 0 {
		return encounter, nil
	}

	round, turn := encounter.Round, encounter.TurnIndex+1
	if turn >= len(encounter.Combatants) {
		turn = 0
		round++
		_ = s.tokens.DecrementConditions(ctx, mapID)
	}
	if err := s.encounters.SetTurn(ctx, encounter.ID, round, turn); err != nil {
		return model.Encounter{}, err
	}
	return s.encounters.Get(ctx, encounter.ID)
}

func (s *CombatService) rollInitiative(ctx context.Context, token model.Token) (int, error) {
	modifier := 0
	if token.CharacterID != nil {
		character, err := s.characters.Get(ctx, *token.CharacterID)
		if err == nil {
			modifier = model.AbilityModifier(character.Abilities.Dexterity)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	} else if token.MonsterID != nil {
		monster, err := s.monsters.Get(ctx, *token.MonsterID)
		if err == nil {
			modifier = model.AbilityModifier(monster.Abilities.Dexterity)
		} else if !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}

	result, err := EvaluateRoll(fmt.Sprintf("1d20%s", signed(modifier)))
	if err != nil {
		return 0, err
	}
	return result.Total, nil
}

// Conditions

// SetCondition toggles a status on a token with optional duration in rounds and notes.
func (s *CombatService) SetCondition(ctx context.Context, tokenID int64, in model.ConditionInput) (model.Token, error) {
	condition := strings.ToLower(strings.TrimSpace(in.Condition))
	if condition == "" {
		return model.Token{}, invalid("condition", "is required")
	}
	if err := s.tokens.SetConditionWithDuration(ctx, tokenID, condition, in.Active, in.DurationRounds, in.Notes); err != nil {
		return model.Token{}, err
	}
	return s.token(ctx, tokenID)
}

// Attack

// Attack resolves one attack: roll against the target armour class, and on a
// hit roll damage and apply it. Everything is decided here, so a client cannot
// report a hit it did not earn.
func (s *CombatService) Attack(ctx context.Context, user model.User, campaignID int64, req model.AttackRequest) (model.DiceRoll, error) {
	source, err := s.attackSource(ctx, user, req)
	if err != nil {
		return model.DiceRoll{}, err
	}

	target, err := s.token(ctx, req.TargetTokenID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.DiceRoll{}, ErrNotFound
	}
	if err != nil {
		return model.DiceRoll{}, err
	}
	vitals, err := s.vitals(ctx, target)
	if err != nil {
		return model.DiceRoll{}, err
	}

	attack, err := EvaluateRoll(attackExpression(source.attackBonus, req.Advantage, req.Disadvantage))
	if err != nil {
		return model.DiceRoll{}, err
	}

	// The natural d20 is the kept die of the first group, which is the only
	// group this expression builds.
	natural := attack.Groups[0].Total
	critical := natural == 20
	fumble := natural == 1
	hit := critical || (!fumble && attack.Total >= vitals.ArmorClass)

	result := model.AttackResult{
		AttackerName: source.attackerName,
		TargetName:   vitals.Name,
		WeaponName:   source.label,
		AttackRoll:   model.DiceSummary{Expression: attack.Expression, Total: attack.Total, Detail: attack},
		TargetAC:     vitals.ArmorClass,
		Hit:          hit,
		Critical:     critical,
		Fumble:       fumble,
		DamageType:   source.damageType,
	}

	if hit {
		damage, err := s.rollDamage(source.damage, critical)
		if err != nil {
			return model.DiceRoll{}, err
		}

		// Apply damage modifier (resistance / vulnerability / immunity)
		finalDamage := calculateModifiedDamage(damage.Total, source.damageType, vitals)
		damage.Total = finalDamage
		result.Damage = &model.DiceSummary{Expression: damage.Expression, Total: finalDamage, Detail: damage}

		if err := s.applyDamage(ctx, target, -finalDamage); err != nil {
			return model.DiceRoll{}, err
		}
		if vitals, err = s.refreshVitals(ctx, req.TargetTokenID); err != nil {
			return model.DiceRoll{}, err
		}
	}
	result.Target = vitals

	return s.record(ctx, user, campaignID, source, result)
}

// calculateModifiedDamage handles 5e damage resistance (/2), vulnerability (*2), and immunity (0).
func calculateModifiedDamage(rawDamage int, damageType string, v model.Vitals) int {
	if damageType == "" || rawDamage <= 0 {
		return rawDamage
	}
	dtype := strings.ToLower(strings.TrimSpace(damageType))
	if hasTag(v.DamageImmunities, dtype) {
		return 0
	}
	vuln := hasTag(v.DamageVulnerabilities, dtype)
	res := hasTag(v.DamageResistances, dtype)
	if vuln && !res {
		return rawDamage * 2
	}
	if res && !vuln {
		return rawDamage / 2
	}
	return rawDamage
}

func hasTag(csvList, tag string) bool {
	if csvList == "" || tag == "" {
		return false
	}
	for _, item := range strings.Split(csvList, ",") {
		if strings.EqualFold(strings.TrimSpace(item), tag) {
			return true
		}
	}
	return false
}

// attackSource resolves and authorizes where the attack comes from. A character
// weapon needs the caller to own that sheet; a monster action needs the DM.
func (s *CombatService) attackSource(ctx context.Context, user model.User, req model.AttackRequest) (attackSource, error) {
	switch {
	case req.ItemID > 0 && req.ActionID > 0:
		return attackSource{}, invalid("item_id", "give either a weapon or a monster action, not both")

	case req.ItemID > 0:
		item, err := s.characters.Item(ctx, req.ItemID)
		if errors.Is(err, sql.ErrNoRows) {
			return attackSource{}, ErrNotFound
		}
		if err != nil {
			return attackSource{}, err
		}
		if !item.IsWeapon() {
			return attackSource{}, invalid("item_id", "has no damage expression, so it is not a weapon")
		}
		// Attacking with a weapon means owning the sheet it is on.
		attacker, err := s.sheets.Get(ctx, user, item.CharacterID)
		if err != nil {
			return attackSource{}, err
		}
		characterID := item.CharacterID
		return attackSource{
			attackerName: attacker.Name,
			label:        item.Name,
			attackBonus:  item.AttackBonus,
			damage:       item.Damage,
			damageType:   item.DamageType,
			characterID:  &characterID,
		}, nil

	case req.ActionID > 0:
		if !user.IsDM() {
			return attackSource{}, ErrForbidden
		}
		action, err := s.monsters.Action(ctx, req.ActionID)
		if errors.Is(err, sql.ErrNoRows) {
			return attackSource{}, ErrNotFound
		}
		if err != nil {
			return attackSource{}, err
		}
		if !action.IsAttack() {
			return attackSource{}, invalid("action_id", "has no damage expression, so it is not an attack")
		}

		// The attacking token names which instance is acting, and has to be one
		// the stat block actually belongs to.
		attackerName := action.Name
		if req.AttackerTokenID > 0 {
			attacker, err := s.tokens.Get(ctx, req.AttackerTokenID)
			if err != nil {
				return attackSource{}, err
			}
			if attacker.MonsterID == nil || *attacker.MonsterID != action.MonsterID {
				return attackSource{}, invalid("attacker_token_id", "does not use that stat block")
			}
			attackerName = attacker.Name
		}
		return attackSource{
			attackerName: attackerName,
			label:        action.Name,
			attackBonus:  action.AttackBonus,
			damage:       action.Damage,
			damageType:   action.DamageType,
		}, nil

	default:
		return attackSource{}, invalid("item_id", "name a weapon or a monster action")
	}
}

// rollDamage doubles the dice on a critical hit, which is the 5e rule: the
// modifier is not doubled, so the second roll contributes its dice only.
func (s *CombatService) rollDamage(expression string, critical bool) (RollResult, error) {
	damage, err := EvaluateRoll(expression)
	if err != nil {
		return RollResult{}, invalid("damage", "the weapon damage expression is not valid")
	}
	if !critical {
		return damage, nil
	}

	extra, err := EvaluateRoll(expression)
	if err != nil {
		return RollResult{}, err
	}
	damage.Groups = append(damage.Groups, extra.Groups...)
	damage.Total += extra.Total - extra.Modifier
	damage.Expression = damage.Expression + " (critical: dice doubled)"
	return damage, nil
}

// applyDamage sends the delta to the linked sheet when there is one, and to the
// token own hit points otherwise.
func (s *CombatService) applyDamage(ctx context.Context, target model.Token, delta int) error {
	if target.CharacterID != nil {
		return s.characters.ApplyDamage(ctx, *target.CharacterID, delta)
	}
	return s.tokens.Damage(ctx, target.ID, delta)
}

// vitals resolves what an attack needs to know about a target: the linked sheet
// wins, and the token own columns are the fallback for a monster.
func (s *CombatService) vitals(ctx context.Context, token model.Token) (model.Vitals, error) {
	v := model.Vitals{
		TokenID:     token.ID,
		CharacterID: token.CharacterID,
		Name:        token.Name,
		HPCurrent:   token.HPCurrent,
		HPMax:       token.HPMax,
		ArmorClass:  token.ArmorClass,
	}

	if token.CharacterID != nil {
		character, err := s.characters.Get(ctx, *token.CharacterID)
		if errors.Is(err, sql.ErrNoRows) {
			return v, nil
		}
		if err != nil {
			return model.Vitals{}, err
		}
		v.Name = character.Name
		v.HPCurrent = character.HPCurrent
		v.HPMax = character.HPMax
		v.HPTemp = character.HPTemp
		v.ArmorClass = character.ArmorClass
		v.DamageResistances = character.DamageResistances
		v.DamageVulnerabilities = character.DamageVulnerabilities
		v.DamageImmunities = character.DamageImmunities
		v.ConditionImmunities = character.ConditionImmunities
		return v, nil
	}

	if token.MonsterID != nil {
		monster, err := s.monsters.Get(ctx, *token.MonsterID)
		if err == nil {
			v.DamageResistances = monster.DamageResistances
			v.DamageVulnerabilities = monster.DamageVulnerabilities
			v.DamageImmunities = monster.DamageImmunities
			v.ConditionImmunities = monster.ConditionImmunities
		}
	}
	return v, nil
}

func (s *CombatService) refreshVitals(ctx context.Context, tokenID int64) (model.Vitals, error) {
	token, err := s.token(ctx, tokenID)
	if err != nil {
		return model.Vitals{}, err
	}
	return s.vitals(ctx, token)
}

// token reads a token with its conditions attached.
func (s *CombatService) token(ctx context.Context, id int64) (model.Token, error) {
	token, err := s.tokens.Get(ctx, id)
	if err != nil {
		return model.Token{}, err
	}
	list := []model.Token{token}
	if err := s.tokens.LoadConditions(ctx, list); err != nil {
		return model.Token{}, err
	}
	return list[0], nil
}

// record stores the resolution as a roll so it lands in the shared feed and
// survives a reconnect like every other roll.
func (s *CombatService) record(ctx context.Context, user model.User, campaignID int64, source attackSource, result model.AttackResult) (model.DiceRoll, error) {
	detail, err := json.Marshal(struct {
		Kind string `json:"kind"`
		model.AttackResult
	}{Kind: RollKindAttack, AttackResult: result})
	if err != nil {
		return model.DiceRoll{}, err
	}

	label := fmt.Sprintf("%s vs %s", source.label, result.TargetName)
	total := result.AttackRoll.Total
	if result.Damage != nil {
		total = result.Damage.Total
	}

	return s.rolls.Record(ctx, user, campaignID, model.DiceRoll{
		CharacterID: source.characterID,
		Label:       label,
		Expression:  result.AttackRoll.Expression,
		Detail:      detail,
		Total:       total,
	})
}

func attackExpression(bonus int, advantage, disadvantage bool) string {
	die := "1d20"
	switch {
	case advantage && !disadvantage:
		die = "2d20kh1"
	case disadvantage && !advantage:
		die = "2d20kl1"
	}
	return die + signed(bonus)
}

func signed(value int) string {
	if value >= 0 {
		return fmt.Sprintf("+%d", value)
	}
	return fmt.Sprintf("%d", value)
}
