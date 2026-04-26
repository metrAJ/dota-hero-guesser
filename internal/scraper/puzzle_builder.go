package scraper

import (
	"doproj/internal/models"

	"github.com/lib/pq"
)

const (
	gameModeTurbo                   = 23
	minimumTurboGameDurationSeconds = 30 * 60 // 30m
)

func extractValidMatches(getheredMatches []models.PublicMatch, processedMatches map[uint64]bool) []models.PublicMatch {
	validMatches := make([]models.PublicMatch, 0, len(getheredMatches)/2)
	for _, m := range getheredMatches {
		if processedMatches[m.MatchID] {
			continue
		}
		if m.Duration < minimumTurboGameDurationSeconds && m.GameMode != gameModeTurbo {
			continue
		}
		validMatches = append(validMatches, m)
	}
	return validMatches
}

func extractPuzzles(match *models.MatchDetailsResponse) []models.Puzzle {
	var validPuzzles []models.Puzzle
	for _, p := range match.Players {
		var mainItemIDs []int64
		for _, itemID := range []int64{p.Item0, p.Item1, p.Item2, p.Item3, p.Item4, p.Item5} {
			if itemID != 0 {
				mainItemIDs = append(mainItemIDs, itemID)
			}
		}
		// Must be at least 6 main items
		if len(mainItemIDs) < 6 {
			continue
		}

		var backpackItemIDs []int64
		for _, itemID := range []int64{p.BackPack0, p.BackPack1, p.BackPack2} {
			if itemID != 0 {
				backpackItemIDs = append(backpackItemIDs, itemID)
			}
		}
		// Check if the player won the match with these items
		isRadiant := p.PlayerSlot < 128
		playerWon := isRadiant == match.RadiantWin
		// Fill the puzzle struct and add to the list of valid puzzles
		validPuzzles = append(validPuzzles, models.Puzzle{
			MatchID:     match.MatchID,
			HeroID:      p.HeroID,
			ItemIDs:     pq.Int64Array(mainItemIDs),
			BackPackIDs: pq.Int64Array(backpackItemIDs),
			IsWon:       playerWon,
		})
	}
	return validPuzzles
}
