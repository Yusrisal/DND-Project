package service

import "testing"

func TestEvaluateRollBounds(t *testing.T) {
	cases := []struct {
		expression string
		min, max   int
	}{
		{"1d20", 1, 20},
		{"1d20+5", 6, 25},
		{"2d6+1d4+3", 6, 23},
		{"4d6kh3", 3, 18},
		{"2d20kh1", 1, 20},
		{"2d20kl1", 1, 20},
		{"d8", 1, 8},
		{"1d4-1", 0, 3},
		{"10", 10, 10},
	}

	for _, c := range cases {
		for range 200 {
			result, err := EvaluateRoll(c.expression)
			if err != nil {
				t.Fatalf("%s: unexpected error: %v", c.expression, err)
			}
			if result.Total < c.min || result.Total > c.max {
				t.Fatalf("%s: total %d outside [%d, %d]", c.expression, result.Total, c.min, c.max)
			}
		}
	}
}

func TestEvaluateRollKeepsTheRightDice(t *testing.T) {
	for range 200 {
		result, err := EvaluateRoll("4d6kh3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		group := result.Groups[0]
		if len(group.Rolls) != 4 {
			t.Fatalf("expected 4 dice, got %d", len(group.Rolls))
		}
		if len(group.Dropped) != 1 {
			t.Fatalf("expected 1 dropped die, got %d", len(group.Dropped))
		}
		for _, roll := range group.Rolls {
			if roll < group.Dropped[0] {
				t.Fatalf("dropped %d but kept a lower die %d", group.Dropped[0], roll)
			}
		}
	}
}

func TestEvaluateRollRejectsNonsense(t *testing.T) {
	for _, expression := range []string{"", "d", "1d1", "0d6", "1000d6", "1d99999", "2d6+", "4d6kh9", "abc"} {
		if _, err := EvaluateRoll(expression); err == nil {
			t.Fatalf("%q: expected an error", expression)
		}
	}
}
