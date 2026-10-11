package model

import "time"

// ClassSkill defines a node in a class skill tree or an ability taught via scroll.
type ClassSkill struct {
	ID             int64     `json:"id"`
	ClassName      string    `json:"class_name"`
	Name           string    `json:"name"`
	Tier           int       `json:"tier"`
	ReqLevel       int       `json:"req_level"`
	MaxRank        int       `json:"max_rank"`
	ParentSkillID  *int64    `json:"parent_skill_id"`
	Damage         string    `json:"damage"`
	DamageType     string    `json:"damage_type"`
	CooldownRounds int       `json:"cooldown_rounds"`
	CostResource   string    `json:"cost_resource"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ClassSkillInput struct {
	ClassName      string `json:"class_name"`
	Name           string `json:"name"`
	Tier           int    `json:"tier"`
	ReqLevel       int    `json:"req_level"`
	MaxRank        int    `json:"max_rank"`
	ParentSkillID  *int64 `json:"parent_skill_id"`
	Damage         string `json:"damage"`
	DamageType     string `json:"damage_type"`
	CooldownRounds int    `json:"cooldown_rounds"`
	CostResource   string `json:"cost_resource"`
	Description    string `json:"description"`
}

// CharacterLearnedSkill tracks an unlocked skill, its current rank, and active status.
type CharacterLearnedSkill struct {
	ID          int64      `json:"id"`
	CharacterID int64      `json:"character_id"`
	SkillID     int64      `json:"skill_id"`
	CurrentRank int        `json:"current_rank"`
	IsActive    bool       `json:"is_active"`
	LearnedAt   time.Time  `json:"learned_at"`
	Skill       ClassSkill `json:"skill"`
}

// PendingLootDrop is a monster drop awaiting DM review (approval or rejection with reason).
type PendingLootDrop struct {
	ID                    int64     `json:"id"`
	CampaignID            int64     `json:"campaign_id"`
	MapID                 int64     `json:"map_id"`
	MonsterName           string    `json:"monster_name"`
	ItemName              string    `json:"item_name"`
	ItemID                *int64    `json:"item_id"`
	Quantity              int       `json:"quantity"`
	CP                    int       `json:"cp"`
	SP                    int       `json:"sp"`
	GP                    int       `json:"gp"`
	Status                string    `json:"status"` // 'pending', 'approved', 'rejected'
	RejectReason          string    `json:"reject_reason"`
	AssignedToCharacterID *int64    `json:"assigned_to_character_id"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// ItemRequirements specifies level, class, and ability score thresholds to equip.
type ItemRequirements struct {
	ReqLevel int    `json:"req_level"`
	ReqClass string `json:"req_class"`
	ReqStr   int    `json:"req_str"`
	ReqDex   int    `json:"req_dex"`
	ReqCon   int    `json:"req_con"`
	ReqInt   int    `json:"req_int"`
	ReqWis   int    `json:"req_wis"`
	ReqCha   int    `json:"req_cha"`
}
