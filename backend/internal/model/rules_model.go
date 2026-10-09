package model

import "slices"

// Ability is one of the six scores, used as the key a skill or save keys off.
type Ability string

const (
	STR Ability = "strength"
	DEX Ability = "dexterity"
	CON Ability = "constitution"
	INT Ability = "intelligence"
	WIS Ability = "wisdom"
	CHA Ability = "charisma"
)

// Skills maps each 5e skill to the ability it uses. The order of SkillNames is
// the order the sheet renders, which is alphabetical as the rulebook prints it.
var Skills = map[string]Ability{
	"acrobatics":      DEX,
	"animal_handling": WIS,
	"arcana":          INT,
	"athletics":       STR,
	"deception":       CHA,
	"history":         INT,
	"insight":         WIS,
	"intimidation":    CHA,
	"investigation":   INT,
	"medicine":        WIS,
	"nature":          INT,
	"perception":      WIS,
	"performance":     CHA,
	"persuasion":      CHA,
	"religion":        INT,
	"sleight_of_hand": DEX,
	"stealth":         DEX,
	"survival":        WIS,
}

var SkillNames = []string{
	"acrobatics", "animal_handling", "arcana", "athletics", "deception",
	"history", "insight", "intimidation", "investigation", "medicine",
	"nature", "perception", "performance", "persuasion", "religion",
	"sleight_of_hand", "stealth", "survival",
}

var SaveNames = []string{
	string(STR), string(DEX), string(CON), string(INT), string(WIS), string(CHA),
}

// Conditions is the 5e list. Statuses are validated against it so the board
// cannot accumulate typos that no client knows how to render.
var Conditions = []string{
	"blinded", "charmed", "deafened", "exhaustion", "frightened", "grappled",
	"incapacitated", "invisible", "paralyzed", "petrified", "poisoned",
	"prone", "restrained", "stunned", "unconscious",
}

func IsCondition(name string) bool { return slices.Contains(Conditions, name) }

// Score reads one ability off the set by name.
func (a Abilities) Score(ability Ability) int {
	switch ability {
	case STR:
		return a.Strength
	case DEX:
		return a.Dexterity
	case CON:
		return a.Constitution
	case INT:
		return a.Intelligence
	case WIS:
		return a.Wisdom
	case CHA:
		return a.Charisma
	}
	return 10
}

// Proficiency records that a character is trained in one skill or save.
// Expertise doubles the bonus, which is how rogues and bards work.
type Proficiency struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Expertise bool   `json:"expertise"`
}

// CheckModifier is what a skill or save check adds: the ability modifier, plus
// the proficiency bonus when trained, doubled with expertise.
type CheckModifier struct {
	Name     string  `json:"name"`
	Ability  Ability `json:"ability"`
	Modifier int     `json:"modifier"`
	Trained  bool    `json:"trained"`
}

// Checks derives every skill and save modifier for a character, so the sheet
// renders a roll button per line and no client reimplements the rule.
func Checks(c Character, proficiencies []Proficiency) (skills, saves []CheckModifier) {
	trained := map[string]Proficiency{}
	for _, p := range proficiencies {
		trained[p.Kind+":"+p.Name] = p
	}

	bonus := func(kind, name string, base int) CheckModifier {
		modifier := base
		p, ok := trained[kind+":"+name]
		if ok {
			modifier += c.ProficiencyBonus
			if p.Expertise {
				modifier += c.ProficiencyBonus
			}
		}
		return CheckModifier{Name: name, Modifier: modifier, Trained: ok}
	}

	for _, name := range SkillNames {
		ability := Skills[name]
		check := bonus("skill", name, AbilityModifier(c.Abilities.Score(ability)))
		check.Ability = ability
		skills = append(skills, check)
	}
	for _, name := range SaveNames {
		ability := Ability(name)
		check := bonus("save", name, AbilityModifier(c.Abilities.Score(ability)))
		check.Ability = ability
		saves = append(saves, check)
	}
	return skills, saves
}
