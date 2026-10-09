package model

import "testing"

func TestAbilityModifier(t *testing.T) {
	// The 5e table, including the negative half where Go integer division
	// rounds toward zero rather than down.
	cases := map[int]int{
		1: -5, 2: -4, 3: -4, 7: -2, 8: -1, 9: -1,
		10: 0, 11: 0, 12: 1, 15: 2, 18: 4, 20: 5, 30: 10,
	}

	for score, want := range cases {
		if got := AbilityModifier(score); got != want {
			t.Errorf("AbilityModifier(%d) = %d, want %d", score, got, want)
		}
	}
}
