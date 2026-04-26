package game

import (
	"context"
	"doproj/internal/models"
	"errors"
)

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

type gameService struct {
	repo GameRepository
}

func NewGameService(repo GameRepository) *gameService {
	return &gameService{repo: repo}
}

// Getting User's game and
func (s *gameService) GetGameView(ctx context.Context, userID uint) (*models.GameStateResponse, error) {
	history, puzzle, err := s.repo.GetGameRound(ctx, userID)
	if err != nil {
		return nil, err
	}
	return s.buildGameResponse(ctx, history, puzzle, "Guess the Hero!")
}

func (s *gameService) ProcessGuess(ctx context.Context, userID uint, roundID uint, guessID uint) (*models.GameStateResponse, error) {
	history, puzzle, err := s.repo.GetActiveRoundByID(ctx, userID, roundID)
	if err != nil || history.Status != models.StatusPlaying {
		return nil, errors.New("invalid or completed round")
	}
	var message string
	if guessID == puzzle.HeroID {
		s.repo.CompleteRound(ctx, userID, history.ID, models.StatusWon)
		history.Status = models.StatusWon
		message = "Correct! You WON!"
	} else {
		newAttempts := history.Attempts + 1
		switch newAttempts {
		case 1:
			message = "Wrong! Here is Dota match outcome."
		case 2:
			message = "Wrong! Here is Backpack Items."
		case 3:
			message = "Wrong! Here is Hero Attribute."
		default:
			s.repo.CompleteRound(ctx, userID, history.ID, models.StatusLost)
			history.Status = models.StatusLost
			message = "Wrong! You LOST!"
		}
		if s.repo.IncrementAttempts(ctx, history.ID) != nil {
			return nil, errors.New("failed to increment attempts")
		}
		history.Attempts++
	}
	return s.buildGameResponse(ctx, history, puzzle, message)
}

// Helper for building JSON
func (s *gameService) buildGameResponse(ctx context.Context, history *models.UserHistory, puzzle *models.Puzzle, message string) (*models.GameStateResponse, error) {

	mainItems, _ := s.repo.GetItemsByIDs(ctx, puzzle.ItemIDs)

	response := &models.GameStateResponse{
		RoundID:   history.ID,
		Attempts:  history.Attempts,
		Status:    history.Status,
		Message:   message,
		MainItems: mainItems,
	}

	// Hint 1: Match Outcome
	if history.Attempts >= 1 {
		response.MatchWon = &puzzle.IsWon
	}

	// Hint 2: Backpack Items
	if history.Attempts >= 2 {
		response.BackpackItems, _ = s.repo.GetItemsByIDs(ctx, puzzle.BackPackIDs)
	}

	// Hint 3: Hero Attribute
	if history.Attempts >= 3 {
		hero, _ := s.repo.GetHeroByID(ctx, puzzle.HeroID)
		response.HeroAttribute = hero.Type
	}

	// Game Ended: Reveal the Hero
	if history.Status != models.StatusPlaying {
		hero, _ := s.repo.GetHeroByID(ctx, puzzle.HeroID)
		response.CorrectHero = hero.Name
	}

	return response, nil
}

// Getting all heroes
func (s *gameService) GetAllHeroes(ctx context.Context) ([]models.Hero, error) {
	return s.repo.GetAllHeroes()
}
