package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// A nil slice marshals to JSON null, and the browser then crashes on
// `.length` of null. Every list the sheet hands to the client has to be an
// empty array instead. This is the bug that took the board down once.
func TestCharacterSheetListsAreNeverNull(t *testing.T) {
	body, err := json.Marshal(CharacterSheet{})
	if err != nil {
		t.Fatalf("marshal sheet: %v", err)
	}

	for _, field := range []string{
		"items", "spells", "skills_loaded", "slots", "proficiencies", "skills", "saves",
	} {
		if strings.Contains(string(body), `"`+field+`":null`) {
			t.Errorf("%s serialized as null, want []", field)
		}
	}
}

func TestConsumableDetection(t *testing.T) {
	skillID := int64(7)

	cases := []struct {
		item   CharacterItem
		potion bool
		portal bool
		combat bool
	}{
		{item: CharacterItem{Name: "Greater Health Potion"}, potion: true},
		{item: CharacterItem{Name: "Elixir of Restoration"}, potion: true},
		{item: CharacterItem{Name: "Town Portal Scroll"}, portal: true},
		{item: CharacterItem{Name: "Dungeon Escape Scroll"}, portal: true},
		{item: CharacterItem{Name: "Scroll of Fireball (1-use)", Damage: "8d6"}, combat: true},
		// A learning scroll teaches a skill, so it is not a one-shot spell.
		{item: CharacterItem{Name: "Scroll of Frost Nova", Damage: "4d6", TeachesSkillID: &skillID}},
		// Plain gear is none of the above.
		{item: CharacterItem{Name: "Traveler Set Helm"}},
	}

	for _, c := range cases {
		if got := c.item.IsPotion(); got != c.potion {
			t.Errorf("%q IsPotion = %v, want %v", c.item.Name, got, c.potion)
		}
		if got := c.item.IsPortalScroll(); got != c.portal {
			t.Errorf("%q IsPortalScroll = %v, want %v", c.item.Name, got, c.portal)
		}
		if got := c.item.IsCombatScroll(); got != c.combat {
			t.Errorf("%q IsCombatScroll = %v, want %v", c.item.Name, got, c.combat)
		}
	}
}
