package model

import (
	"encoding/json"
	"time"
)

type Map struct {
	ID          int64     `json:"id"`
	CampaignID  int64     `json:"campaign_id"`
	Name        string    `json:"name"`
	ImageURL    string    `json:"image_url"`
	GridSize    int       `json:"grid_size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	FogEnabled  bool      `json:"fog_enabled"`
	FogRevealed string    `json:"fog_revealed"` // JSON string array of revealed regions
	Weather     string    `json:"weather"`      // clear, rain, snow, embers, fog
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MapInput struct {
	CampaignID  int64  `json:"campaign_id"`
	Name        string `json:"name"`
	ImageURL    string `json:"image_url"`
	GridSize    int    `json:"grid_size"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FogEnabled  bool   `json:"fog_enabled"`
	FogRevealed string `json:"fog_revealed"`
	Weather     string `json:"weather"`
}

type FogInput struct {
	FogEnabled  bool   `json:"fog_enabled"`
	FogRevealed string `json:"fog_revealed"`
}

type WeatherInput struct {
	Weather string `json:"weather"`
}

type AoETemplate struct {
	ID        int64     `json:"id"`
	MapID     int64     `json:"map_id"`
	Name      string    `json:"name"`
	Shape     string    `json:"shape"` // circle, cone, line, cube
	SizeFt    int       `json:"size_ft"`
	Color     string    `json:"color"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	CreatedAt time.Time `json:"created_at"`
}

type AoETemplateInput struct {
	MapID  int64  `json:"map_id"`
	Name   string `json:"name"`
	Shape  string `json:"shape"`
	SizeFt int    `json:"size_ft"`
	Color  string `json:"color"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

type MapObject struct {
	ID        int64           `json:"id"`
	MapID     int64           `json:"map_id"`
	Kind      string          `json:"kind"` // door, chest, trap, marker
	Name      string          `json:"name"`
	X         int             `json:"x"`
	Y         int             `json:"y"`
	State     string          `json:"state"` // open, closed, locked, looted, triggered, hidden
	Data      json.RawMessage `json:"data"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type MapObjectInput struct {
	MapID int64           `json:"map_id"`
	Kind  string          `json:"kind"`
	Name  string          `json:"name"`
	X     int             `json:"x"`
	Y     int             `json:"y"`
	State string          `json:"state"`
	Data  json.RawMessage `json:"data"`
}

type MapPing struct {
	MapID      int64   `json:"map_id"`
	X          float64 `json:"x"`
	Y          float64 `json:"y"`
	AuthorName string  `json:"author_name"`
	Color      string  `json:"color"`
}
