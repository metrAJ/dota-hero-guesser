package repository

import (
	"context"
	"doproj/internal/models"
)

type ScraperRepository interface {
	SaveHero(hero *models.Hero) error
	SaveItem(item *models.Item) error
	SavePuzzle(puzzle *models.Puzzle) error
	GetProcessedMatchIDs(ctx context.Context) (map[uint64]bool, error)
}

type GameRepository interface {
	GetUnplayedPuzzle(ctx context.Context, userId uint) (*models.Puzzle, error)
	GetGameRound(ctx context.Context, userID uint) (*models.UserHistory, *models.Puzzle, error)
	IncrementAttempts(ctx context.Context, roundID uint) error
	GetItemsByIDs(ctx context.Context, itemIDs []int64) ([]models.Item, error)
	GetHeroByID(ctx context.Context, heroID uint) (*models.Hero, error)
	GetActiveRoundByID(ctx context.Context, userID uint, roundID uint) (*models.UserHistory, *models.Puzzle, error)
	CompleteRound(ctx context.Context, userID uint, roundID uint, finalStatus string) error
	GetAllHeroes() ([]models.Hero, error)
}

type UserRepository interface {
	GetTopPlayers(ctx context.Context) ([]models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByID(ctx context.Context, userID uint) (*models.User, error)
}

type MultiplayerRepository interface {
	CreateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error
	UpdateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error
	FindUnplayedMutualPuzzle(ctx context.Context, p1ID, p2ID uint) (*models.Puzzle, error)
}
