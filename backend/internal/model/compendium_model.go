package model

import "time"

type Environment struct {
	ID                int64     `json:"id"`
	CampaignID        int64     `json:"campaign_id"`
	MapID             *int64    `json:"map_id"`
	Name              string    `json:"name"`
	Kind              string    `json:"kind"` // hazard, weather, terrain, trap, aura
	DifficultyDC      int       `json:"difficulty_dc"`
	Damage            string    `json:"damage"`
	DamageType        string    `json:"damage_type"`
	EffectDescription string    `json:"effect_description"`
	Active            bool      `json:"active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type EnvironmentInput struct {
	CampaignID        int64  `json:"campaign_id"`
	MapID             *int64 `json:"map_id"`
	Name              string `json:"name"`
	Kind              string `json:"kind"`
	DifficultyDC      int    `json:"difficulty_dc"`
	Damage            string `json:"damage"`
	DamageType        string `json:"damage_type"`
	EffectDescription string `json:"effect_description"`
	Active            bool   `json:"active"`
}

type CompendiumItem struct {
	ID             int64     `json:"id"`
	CampaignID     *int64    `json:"campaign_id"`
	Name           string    `json:"name"`
	Category       string    `json:"category"` // weapon, armor, potion, scroll, wondrous, gear
	Rarity         string    `json:"rarity"`   // common, uncommon, rare, very rare, legendary
	CostCP         int       `json:"cost_cp"`
	Weight         float64   `json:"weight"`
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
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CompendiumItemInput struct {
	CampaignID     *int64  `json:"campaign_id"`
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Rarity         string  `json:"rarity"`
	CostCP         int     `json:"cost_cp"`
	Weight         float64 `json:"weight"`
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
}

type MonsterDrop struct {
	ID          int64     `json:"id"`
	MonsterID   int64     `json:"monster_id"`
	ItemName    string    `json:"item_name"`
	ItemID      *int64    `json:"item_id"`
	DropRate    int       `json:"drop_rate"` // 1 to 100 percentage
	MinQuantity int       `json:"min_quantity"`
	MaxQuantity int       `json:"max_quantity"`
	CP          int       `json:"cp"`
	SP          int       `json:"sp"`
	GP          int       `json:"gp"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MonsterDropInput struct {
	MonsterID   int64  `json:"monster_id"`
	ItemName    string `json:"item_name"`
	ItemID      *int64 `json:"item_id"`
	DropRate    int    `json:"drop_rate"`
	MinQuantity int    `json:"min_quantity"`
	MaxQuantity int    `json:"max_quantity"`
	CP          int    `json:"cp"`
	SP          int    `json:"sp"`
	GP          int    `json:"gp"`
}

type CompendiumSpell struct {
	ID          int64     `json:"id"`
	CampaignID  *int64    `json:"campaign_id"`
	Name        string    `json:"name"`
	Level       int       `json:"level"`
	School      string    `json:"school"`
	CastingTime string    `json:"casting_time"`
	RangeArea   string    `json:"range_area"`
	Duration    string    `json:"duration"`
	Damage      string    `json:"damage"`
	DamageType  string    `json:"damage_type"`
	SaveAbility string    `json:"save_ability"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CompendiumSpellInput struct {
	CampaignID  *int64 `json:"campaign_id"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	School      string `json:"school"`
	CastingTime string `json:"casting_time"`
	RangeArea   string `json:"range_area"`
	Duration    string `json:"duration"`
	Damage      string `json:"damage"`
	DamageType  string `json:"damage_type"`
	SaveAbility string `json:"save_ability"`
	Description string `json:"description"`
}

type StatusEffect struct {
	ID                    int64     `json:"id"`
	CampaignID            *int64    `json:"campaign_id"`
	Name                  string    `json:"name"`
	Kind                  string    `json:"kind"` // condition, buff, debuff, hazard
	DefaultDurationRounds int       `json:"default_duration_rounds"`
	Color                 string    `json:"color"`
	EffectDescription     string    `json:"effect_description"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type StatusEffectInput struct {
	CampaignID            *int64 `json:"campaign_id"`
	Name                  string `json:"name"`
	Kind                  string `json:"kind"`
	DefaultDurationRounds int    `json:"default_duration_rounds"`
	Color                 string `json:"color"`
	EffectDescription     string `json:"effect_description"`
}

type DroppedItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type LootRollResult struct {
	MonsterName string        `json:"monster_name"`
	Items       []DroppedItem `json:"items"`
	CP          int           `json:"cp"`
	SP          int           `json:"sp"`
	GP          int           `json:"gp"`
}
