package models

import (
	"github.com/lib/pq"
	"gorm.io/gorm"
)

const (
	StatusPlaying = "playing"
	StatusWon     = "won"
	StatusLost    = "lost"
)

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Wins     int    `gorm:"default:0" json:"wins"`
	Losses   int    `gorm:"default:0" json:"losses"`
}

type UserHistory struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	UserID   uint   `gorm:"index:idx_active_game" json:"user_id"`
	User     User   `gorm:"foreignKey:UserID"`
	PuzzleID uint   `json:"puzzle_id"`
	Puzzle   Puzzle `gorm:"foreignKey:PuzzleID"`
	Attempts int    `gorm:"default:0" json:"attempts"`
	Status   string `gorm:"index:idx_active_game;default:'playing'" json:"status"`
}

type MultiplayerMatch struct {
	gorm.Model
	Player1ID uint `json:"player1_id"`
	Player1   User `gorm:"foreignKey:Player1ID" json:"player1"`

	Player2ID uint `json:"player2_id"`
	Player2   User `gorm:"foreignKey:Player2ID" json:"player2"`

	PuzzleID uint   `json:"puzzle_id"`
	Puzzle   Puzzle `gorm:"foreignKey:PuzzleID" json:"puzzle"`
	Status   string `gorm:"default:'playing'" json:"status"`
	WinnerId uint   `json:"winner_id"`
}

type Item struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
}

type Hero struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	Type     string `json:"type"` //strength, agility, intelligence, universal
}

type PublicMatch struct {
	MatchID  uint64 `json:"match_id"`
	Duration int    `json:"duration"`
	GameMode int    `json:"game_mode"`
}

type Puzzle struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	MatchID     uint64        `json:"match_id"`
	HeroID      uint          `json:"hero_id"`
	Hero        Hero          `gorm:"foreignKey:HeroID"`
	ItemIDs     pq.Int64Array `gorm:"type:integer[]" json:"item_ids"`
	BackPackIDs pq.Int64Array `gorm:"type:integer[]" json:"backpack_ids"`
	IsWon       bool          `gorm:"not null" json:"is_won"`
}

type GameStateResponse struct {
	RoundID  uint   `json:"round_id"`
	Attempts int    `json:"attempts"`
	Message  string `json:"message"`
	Status   string `json:"status"`

	MainItems     []Item `json:"main_items,omitempty"`
	MatchWon      *bool  `json:"is_won,omitempty"`
	BackpackItems []Item `json:"backpack_items"`
	HeroAttribute string `json:"hero_attribute,omitempty"`
	CorrectHero   string `json:"correct_hero,omitempty"`
}

type UserStatsResponse struct {
	Name   string `json:"name"`
	Wins   int    `json:"wins"`
	Losses int    `json:"losses"`
}

type MatchDetailsResponse struct {
	MatchID    uint64        `json:"match_id"`
	RadiantWin bool          `json:"radiant_win"`
	Players    []MatchPlayer `json:"players"`
}

type MatchPlayer struct {
	HeroID     uint  `json:"hero_id"`
	PlayerSlot uint  `json:"player_slot"`
	Item0      int64 `json:"item_0"`
	Item1      int64 `json:"item_1"`
	Item2      int64 `json:"item_2"`
	Item3      int64 `json:"item_3"`
	Item4      int64 `json:"item_4"`
	Item5      int64 `json:"item_5"`
	BackPack0  int64 `json:"backpack_0"`
	BackPack1  int64 `json:"backpack_1"`
	BackPack2  int64 `json:"backpack_2"`
}
