package main

import (
	"context"
	"doproj/internal/config"
	"doproj/internal/database"
	"doproj/internal/repository"
	"doproj/internal/scraper"
	"doproj/pkg/opendota"
	"flag"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	// flags for console -getheroes | -getitems | -getgames=N
	updateHeroes := flag.Bool("getheroes", false, "Update hero data")
	udateItems := flag.Bool("getitems", false, "Update item data")
	addPuzzle := flag.Int("getgames", 0, "Number of recent games to parse into puzzles")
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Initialize database connection -> Repository -> Scraper
	db := database.InitDB(cfg)
	repo := repository.NewGormRepo(db)
	dotaScraper := scraper.NewScraper(repo, opendota.NewOpenDotaClient())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	if *updateHeroes {
		err = dotaScraper.SeedHeroes(ctx)
	}
	if *udateItems {
		err = dotaScraper.SeedItems(ctx)
	}
	if *addPuzzle > 0 {
		err = dotaScraper.SeedPuzzles(ctx, *addPuzzle)
	}
	if err != nil {
		fmt.Println(err)
	}
	cancel()
}
