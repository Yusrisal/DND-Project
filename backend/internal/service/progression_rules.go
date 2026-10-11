package service

import "math"

// XPTableLevelCap defines the maximum level supported in the progression system.
const XPTableLevelCap = 50

// LevelThresholdXP returns total XP required to reach a specific level (1-50).
// Levels 1-6 have a smooth early curve. Levels 7-50 scale steeply for long-run play.
func LevelThresholdXP(level int) int64 {
	if level <= 1 {
		return 0
	}
	if level > XPTableLevelCap {
		level = XPTableLevelCap
	}

	// Early game: levels 2 - 6
	earlyXP := []int64{0, 0, 300, 900, 1800, 3200, 5000}
	if level <= 6 {
		return earlyXP[level]
	}

	// Mid/Endgame: Level 7-50 steep curve
	// Base at Lv6 is 5000 XP. Each subsequent level scales exponentially.
	// step = 2000 * (level - 5)^1.85
	var total int64 = 5000
	for lvl := 7; lvl <= level; lvl++ {
		step := int64(math.Round(2200 * math.Pow(float64(lvl-5), 1.82)))
		total += step
	}
	return total
}

// LevelFromXP computes the calculated level given total XP without considering DM gating.
func LevelFromXP(xp int64) int {
	if xp <= 0 {
		return 1
	}
	for lvl := XPTableLevelCap; lvl >= 1; lvl-- {
		if xp >= LevelThresholdXP(lvl) {
			return lvl
		}
	}
	return 1
}
