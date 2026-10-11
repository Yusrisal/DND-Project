package model

import (
	"strings"
	"time"
)

// Abilities are the six 5e scores. They are a separate struct so the modifier
// rule lives in one place instead of six.
type Abilities struct {
	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Constitution int `json:"constitution"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Charisma     int `json:"charisma"`
}

// AbilityModifier is the 5e rule: floor((score - 10) / 2), which has to round
// down for scores below 10 as well, where Go integer division rounds toward zero.
func AbilityModifier(score int) int {
	diff := score - 10
	if diff < 0 {
		return (diff - 1) / 2
	}
	return diff / 2
}

// Modifiers is sent alongside the raw scores so every client does not
// reimplement the rule, and they cannot disagree about it.
func (a Abilities) Modifiers() Abilities {
	return Abilities{
		Strength:     AbilityModifier(a.Strength),
		Dexterity:    AbilityModifier(a.Dexterity),
		Constitution: AbilityModifier(a.Constitution),
		Intelligence: AbilityModifier(a.Intelligence),
		Wisdom:       AbilityModifier(a.Wisdom),
		Charisma:     AbilityModifier(a.Charisma),
	}
}

type Character struct {
	ID                    int64     `json:"id"`
	CampaignID            int64     `json:"campaign_id"`
	UserID                *int64    `json:"user_id"`
	Status                string    `json:"status"`
	Name                  string    `json:"name"`
	Class                 string    `json:"class"`
	Race                  string    `json:"race"`
	Level                 int       `json:"level"`
	XP                    int64     `json:"xp"`
	StatPoints            int       `json:"stat_points"`
	SkillPoints           int       `json:"skill_points"`
	LevelBlocked          bool      `json:"level_blocked"`
	BlockReason           string    `json:"block_reason"`
	HPCurrent             int       `json:"hp_current"`
	HPMax                 int       `json:"hp_max"`
	HPTemp                int       `json:"hp_temp"`
	ArmorClass            int       `json:"armor_class"`
	Speed                 int       `json:"speed"`
	ProficiencyBonus      int       `json:"proficiency_bonus"`
	CP                    int       `json:"cp"`
	SP                    int       `json:"sp"`
	EP                    int       `json:"ep"`
	GP                    int       `json:"gp"`
	PP                    int       `json:"pp"`
	Abilities             Abilities `json:"abilities"`
	Modifiers             Abilities `json:"modifiers"`
	Notes                 string    `json:"notes"`
	DeathSaveSuccesses    int       `json:"death_save_successes"`
	DeathSaveFailures     int       `json:"death_save_failures"`
	HitDiceTotal          int       `json:"hit_dice_total"`
	HitDiceUsed           int       `json:"hit_dice_used"`
	HitDiceDie            string    `json:"hit_dice_die"`
	DamageResistances     string    `json:"damage_resistances"`
	DamageVulnerabilities string    `json:"damage_vulnerabilities"`
	DamageImmunities      string    `json:"damage_immunities"`
	ConditionImmunities   string    `json:"condition_immunities"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type CharacterInput struct {
	CampaignID            int64     `json:"campaign_id"`
	UserID                *int64    `json:"user_id"`
	Status                string    `json:"status"`
	Name                  string    `json:"name"`
	Class                 string    `json:"class"`
	Race                  string    `json:"race"`
	Level                 int       `json:"level"`
	XP                    int64     `json:"xp"`
	StatPoints            int       `json:"stat_points"`
	SkillPoints           int       `json:"skill_points"`
	LevelBlocked          bool      `json:"level_blocked"`
	BlockReason           string    `json:"block_reason"`
	HPCurrent             int       `json:"hp_current"`
	HPMax                 int       `json:"hp_max"`
	HPTemp                int       `json:"hp_temp"`
	ArmorClass            int       `json:"armor_class"`
	Speed                 int       `json:"speed"`
	ProficiencyBonus      int       `json:"proficiency_bonus"`
	CP                    int       `json:"cp"`
	SP                    int       `json:"sp"`
	EP                    int       `json:"ep"`
	GP                    int       `json:"gp"`
	PP                    int       `json:"pp"`
	Abilities             Abilities `json:"abilities"`
	Notes                 string    `json:"notes"`
	DeathSaveSuccesses    int       `json:"death_save_successes"`
	DeathSaveFailures     int       `json:"death_save_failures"`
	HitDiceTotal          int       `json:"hit_dice_total"`
	HitDiceUsed           int       `json:"hit_dice_used"`
	HitDiceDie            string    `json:"hit_dice_die"`
	DamageResistances     string    `json:"damage_resistances"`
	DamageVulnerabilities string    `json:"damage_vulnerabilities"`
	DamageImmunities      string    `json:"damage_immunities"`
	ConditionImmunities   string    `json:"condition_immunities"`
}

type CharacterItem struct {
	ID             int64     `json:"id"`
	CharacterID    int64     `json:"character_id"`
	Name           string    `json:"name"`
	Quantity       int       `json:"quantity"`
	Weight         float64   `json:"weight"`
	CostCP         int       `json:"cost_cp"`
	Equipped       bool      `json:"equipped"`
	AttackBonus    int       `json:"attack_bonus"`
	Damage         string    `json:"damage"`
	DamageType     string    `json:"damage_type"`
	Description    string    `json:"description"`
	ReqLevel       int       `json:"req_level"`
	ReqClass       string    `json:"req_class"`
	ReqStr         int       `json:"req_str"`
	ReqDex         int       `json:"req_dex"`
	ReqCon         int       `json:"req_con"`
	ReqInt         int       `json:"req_int"`
	ReqWis         int       `json:"req_wis"`
	ReqCha         int       `json:"req_cha"`
	TeachesSkillID *int64    `json:"teaches_skill_id"`
	SetName        string    `json:"set_name"`
	SetPiece       string    `json:"set_piece"`
	SetBonus2pc    string    `json:"set_bonus_2pc"`
	SetBonus4pc    string    `json:"set_bonus_4pc"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// IsWeapon is the only thing that marks an inventory row as attackable: it has
// a damage expression.
func (i CharacterItem) IsWeapon() bool { return i.Damage != "" }

// Consumable detection helpers
func (i CharacterItem) IsPotion() bool {
	lower := strings.ToLower(i.Name)
	return strings.Contains(lower, "potion") || strings.Contains(lower, "elixir") || strings.Contains(lower, "brew") || strings.Contains(lower, "flask")
}

func (i CharacterItem) IsPortalScroll() bool {
	lower := strings.ToLower(i.Name)
	return strings.Contains(lower, "portal") || strings.Contains(lower, "teleport") || strings.Contains(lower, "escape scroll")
}

func (i CharacterItem) IsCombatScroll() bool {
	lower := strings.ToLower(i.Name)
	return strings.Contains(lower, "scroll") && i.TeachesSkillID == nil && i.Damage != ""
}

type ConsumeItemResult struct {
	ItemName    string `json:"item_name"`
	EffectType  string `json:"effect_type"` // "heal", "portal", "spell"
	EffectValue string `json:"effect_value"`
	HealAmount  int    `json:"heal_amount,omitempty"`
	Damage      string `json:"damage,omitempty"`
	Message     string `json:"message"`
}

type CharacterItemInput struct {
	Name           string  `json:"name"`
	Quantity       int     `json:"quantity"`
	Weight         float64 `json:"weight"`
	CostCP         int     `json:"cost_cp"`
	Equipped       bool    `json:"equipped"`
	AttackBonus    int     `json:"attack_bonus"`
	Damage         string  `json:"damage"`
	DamageType     string  `json:"damage_type"`
	Description    string  `json:"description"`
	ReqLevel       int     `json:"req_level"`
	ReqClass       string  `json:"req_class"`
	ReqStr         int     `json:"req_str"`
	ReqDex         int     `json:"req_dex"`
	ReqCon         int     `json:"req_con"`
	ReqInt         int     `json:"req_int"`
	ReqWis         int     `json:"req_wis"`
	ReqCha         int     `json:"req_cha"`
	TeachesSkillID *int64  `json:"teaches_skill_id"`
	SetName        string  `json:"set_name"`
	SetPiece       string  `json:"set_piece"`
	SetBonus2pc    string  `json:"set_bonus_2pc"`
	SetBonus4pc    string  `json:"set_bonus_4pc"`
}

// CharacterSpell is a free-text entry rather than a reference into a spell
// catalogue: there is no SRD data in this project to point at.
type CharacterSpell struct {
	ID          int64     `json:"id"`
	CharacterID int64     `json:"character_id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	School      string    `json:"school"`
	Prepared    bool      `json:"prepared"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CharacterSpellInput struct {
	Name        string `json:"name"`
	Level       int    `json:"level"`
	School      string `json:"school"`
	Prepared    bool   `json:"prepared"`
	Description string `json:"description"`
}

// SpellSlot tracks one spell level. Levels run 1 to 9; cantrips are level 0
// spells and have no slots.
type SpellSlot struct {
	Level int `json:"level"`
	Total int `json:"total"`
	Used  int `json:"used"`
}

type SpellSlotInput struct {
	Total int `json:"total"`
	Used  int `json:"used"`
}

// CharacterSheet is the whole sheet in one response, which is what the player
// panel loads.
type CharacterSheet struct {
	Character     Character               `json:"character"`
	Items         []CharacterItem         `json:"items"`
	Spells        []CharacterSpell        `json:"spells"`
	SkillsLoaded  []CharacterLearnedSkill `json:"skills_loaded"`
	Slots         []SpellSlot             `json:"slots"`
	Proficiencies []Proficiency           `json:"proficiencies"`
	// Skills and Saves are derived, so no client has to reimplement the
	// proficiency rule and then disagree with the server about it.
	Skills []CheckModifier `json:"skills"`
	Saves  []CheckModifier `json:"saves"`
}

// TotalCP converts coins to copper piece value (10 CP = 1 SP, 50 CP = 1 EP, 100 CP = 1 GP, 1000 CP = 1 PP).
func (c *Character) TotalCP() int {
	return c.CP + c.SP*10 + c.EP*50 + c.GP*100 + c.PP*1000
}
