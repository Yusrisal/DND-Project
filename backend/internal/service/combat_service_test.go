package service

import (
	"dndproject/backend/internal/model"
	"testing"
)

func TestAttackExpression(t *testing.T) {
	cases := []struct {
		bonus        int
		advantage    bool
		disadvantage bool
		want         string
	}{
		{3, false, false, "1d20+3"},
		{0, false, false, "1d20+0"},
		{-1, false, false, "1d20-1"},
		{5, true, false, "2d20kh1+5"},
		{5, false, true, "2d20kl1+5"},
		// Advantage and disadvantage together cancel, as they do in 5e.
		{2, true, true, "1d20+2"},
	}

	for _, c := range cases {
		if got := attackExpression(c.bonus, c.advantage, c.disadvantage); got != c.want {
			t.Errorf("attackExpression(%d, %t, %t) = %q, want %q", c.bonus, c.advantage, c.disadvantage, got, c.want)
		}
	}
}

func TestRollDamageDoublesDiceNotModifier(t *testing.T) {
	s := &CombatService{}

	for range 200 {
		normal, err := s.rollDamage("1d8+3", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if normal.Total < 4 || normal.Total > 11 {
			t.Fatalf("1d8+3 rolled %d, outside [4, 11]", normal.Total)
		}

		// A critical doubles the dice only: 2d8 + 3, so 5 to 19 rather than 22.
		critical, err := s.rollDamage("1d8+3", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if critical.Total < 5 || critical.Total > 19 {
			t.Fatalf("critical 1d8+3 rolled %d, outside [5, 19]", critical.Total)
		}
		if len(critical.Groups) != 2 {
			t.Fatalf("expected 2 dice groups on a critical, got %d", len(critical.Groups))
		}
	}
}

func TestRollDamageRejectsNonsense(t *testing.T) {
	s := &CombatService{}
	if _, err := s.rollDamage("not dice", false); err == nil {
		t.Fatal("expected an error for an invalid damage expression")
	}
}

func TestCalculateModifiedDamage(t *testing.T) {
	cases := []struct {
		rawDamage  int
		damageType string
		vitals     model.Vitals
		want       int
	}{
		{10, "fire", model.Vitals{}, 10},
		{10, "fire", model.Vitals{DamageResistances: "fire,cold"}, 5},
		{10, "fire", model.Vitals{DamageVulnerabilities: "fire,poison"}, 20},
		{10, "fire", model.Vitals{DamageImmunities: "fire"}, 0},
		{10, "radiant", model.Vitals{DamageResistances: "slashing", DamageImmunities: "necrotic"}, 10},
		{15, "cold", model.Vitals{DamageResistances: "cold"}, 7}, // integer division rounds down
	}

	for _, c := range cases {
		got := calculateModifiedDamage(c.rawDamage, c.damageType, c.vitals)
		if got != c.want {
			t.Errorf("calculateModifiedDamage(%d, %q, %+v) = %d, want %d", c.rawDamage, c.damageType, c.vitals, got, c.want)
		}
	}
}
