package model

import "time"

type Shop struct {
	ID          int64     `json:"id"`
	CampaignID  int64     `json:"campaign_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ShopInput struct {
	CampaignID  int64  `json:"campaign_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ShopItem struct {
	ID          int64     `json:"id"`
	ShopID      int64     `json:"shop_id"`
	Name        string    `json:"name"`
	CostCP      int       `json:"cost_cp"`
	Quantity    int       `json:"quantity"` // -1 = unlimited
	Weight      float64   `json:"weight"`
	AttackBonus int       `json:"attack_bonus"`
	Damage      string    `json:"damage"`
	DamageType  string    `json:"damage_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ShopItemInput struct {
	Name        string  `json:"name"`
	CostCP      int     `json:"cost_cp"`
	Quantity    int     `json:"quantity"`
	Weight      float64 `json:"weight"`
	AttackBonus int     `json:"attack_bonus"`
	Damage      string  `json:"damage"`
	DamageType  string  `json:"damage_type"`
	Description string  `json:"description"`
}

type ShopDetails struct {
	Shop  Shop       `json:"shop"`
	Items []ShopItem `json:"items"`
}

type BuyItemRequest struct {
	CharacterID int64 `json:"character_id"`
	Quantity    int   `json:"quantity"`
}

type SellItemRequest struct {
	CharacterID int64 `json:"character_id"`
	ItemID      int64 `json:"item_id"`
	Quantity    int   `json:"quantity"`
}

type SellItemResult struct {
	ItemName     string `json:"item_name"`
	QuantitySold int    `json:"quantity_sold"`
	RefundCP     int    `json:"refund_cp"`
	CP           int    `json:"cp"`
	SP           int    `json:"sp"`
	EP           int    `json:"ep"`
	GP           int    `json:"gp"`
	PP           int    `json:"pp"`
}

type LootSplitRequest struct {
	CampaignID int64 `json:"campaign_id"`
	CP         int   `json:"cp"`
	SP         int   `json:"sp"`
	EP         int   `json:"ep"`
	GP         int   `json:"gp"`
	PP         int   `json:"pp"`
}
