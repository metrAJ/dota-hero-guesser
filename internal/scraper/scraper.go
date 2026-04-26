package scraper

import (
	"context"
	"doproj/internal/models"
	"fmt"
	"time"
)

type ScraperRepository interface {
	SaveHero(hero *models.Hero) error
	SaveItem(item *models.Item) error
	SavePuzzle(puzzle *models.Puzzle) error
	GetProcessedMatchIDs(ctx context.Context) (map[uint64]bool, error)
}

type DataProvider interface {
	FetchHeroes(ctx context.Context) ([]models.Hero, error)
	FetchItems(ctx context.Context) ([]models.Item, error)
	FetchPublicMatches(ctx context.Context) ([]models.PublicMatch, error)
	FetchMatchDetales(ctx context.Context, matchID uint64) (*models.MatchDetailsResponse, error)
}

type DotaScraper struct {
	repo         ScraperRepository
	dataProvider DataProvider
}

func NewScraper(repo ScraperRepository, dataProvider DataProvider) *DotaScraper {
	return &DotaScraper{
		repo:         repo,
		dataProvider: dataProvider,
	}
}

func (s *DotaScraper) SeedHeroes(ctx context.Context) error {
	heroes, err := s.dataProvider.FetchHeroes(ctx)
	if err != nil {
		return fmt.Errorf("dataProvider.FetchHeroes: %w", err)
	}
	for _, hero := range heroes {
		if err := s.repo.SaveHero(&hero); err != nil {
			fmt.Println("repo.SaveHero: %w", err)
		}
	}
	fmt.Println("Heroes synced")
	return nil
}

func (s *DotaScraper) SeedItems(ctx context.Context) error {
	items, err := s.dataProvider.FetchItems(ctx)
	if err != nil {
		return fmt.Errorf("dataProvider.FetchItems: %w", err)
	}
	for _, item := range items {
		if err := s.repo.SaveItem(&item); err != nil {
			fmt.Println("repo.SaveItem: %w", err)
		}
	}
	fmt.Println("Items synced")
	return nil
}

func (s *DotaScraper) SeedPuzzles(ctx context.Context, limit int) error {
	// Get random matches
	matches, err := s.dataProvider.FetchPublicMatches(ctx)
	if err != nil {
		return fmt.Errorf("dataProvider.FetchItems: %w", err)
	}
	// Get map of already processed matche IDs from DB
	processedMatches, err := s.repo.GetProcessedMatchIDs(ctx)
	if err != nil {
		return fmt.Errorf("repo.GetProcessedMatchIDs: %w", err)
	}
	// Extract unique matches, that satisfy our puzzle goal
	validMatches := extractValidMatches(matches, processedMatches)
	processedGames := 0
	savedPuzzles := 0
	for _, m := range validMatches {
		if processedGames >= limit {
			break
		}
		// Get match detales for puzzles
		matchDetales, err := s.dataProvider.FetchMatchDetales(ctx, m.MatchID)
		if err != nil {
			return fmt.Errorf("dataProvider.FetchMatchDetales: %w", err)
		}
		if matchDetales == nil {
			continue
		}
		// Extract puzzles from match detales
		puzzles := extractPuzzles(matchDetales)
		// Save puzzles
		for _, puzzle := range puzzles {
			err := s.repo.SavePuzzle(&puzzle)
			if err == nil {
				savedPuzzles++
			}
		}
		processedGames++
		time.Sleep(time.Second / 2)
	}
	fmt.Printf("( %d ) Puzzles added", savedPuzzles)
	return nil
}
