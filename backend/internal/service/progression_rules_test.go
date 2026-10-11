package service

import "testing"

func TestLevelFromXP(t *testing.T) {
	cases := []struct {
		xp        int64
		wantLevel int
	}{
		{0, 1},
		{299, 1},
		{300, 2},
		{899, 2},
		{900, 3},
		{1800, 4},
		{3200, 5},
		{5000, 6},
		{5001, 6},
	}

	for _, c := range cases {
		got := LevelFromXP(c.xp)
		if got != c.wantLevel {
			t.Errorf("LevelFromXP(%d) = %d, want %d", c.xp, got, c.wantLevel)
		}
	}
}

func TestLevel50MaxThreshold(t *testing.T) {
	xp50 := LevelThresholdXP(50)
	if xp50 <= LevelThresholdXP(49) {
		t.Errorf("Level 50 threshold (%d) should be greater than Level 49 (%d)", xp50, LevelThresholdXP(49))
	}
	if LevelFromXP(xp50) != 50 {
		t.Errorf("LevelFromXP(%d) expected 50, got %d", xp50, LevelFromXP(xp50))
	}
}
