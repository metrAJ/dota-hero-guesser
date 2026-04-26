package repository

import (
	"context"
	"doproj/internal/config"
	"doproj/internal/database"
	"doproj/internal/models"
	"testing"
)

func TestGetItemsByIDs(t *testing.T) {
	ctx := context.Background()
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	db := database.InitDB(cfg)
	repo := NewGormRepo(db)
	itemIDs := []int64{108, 34, 77}
	items, err := repo.GetItemsByIDs(ctx, itemIDs)

	if err != nil {
		t.Fatalf("Got error : %v", err)
	}
	if len(items) != len(itemIDs) {
		t.Errorf("Expected %d items, got %d", len(itemIDs), len(items))
	}
	t.Logf("Successfully fetched items:")
	for _, item := range items {
		t.Logf(" - ID: %d, Name: %s", item.ID, item.Name)
	}

}

func TestGameFlow(t *testing.T) {
	ctx := context.Background()
	//Conection
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	db := database.InitDB(cfg)
	repo := NewGormRepo(db)
	// Temp user
	testUser := models.User{
		Username: "test_user",
		Password: "test_password",
	}
	// Delete if prev test crashed
	db.Where("username = ?", testUser.Username).Delete(&models.User{})
	// Create tets user
	if err := db.Create(&testUser).Error; err != nil {
		t.Fatalf("Could not create test user: %v", err)
	}
	// Cleaning
	defer func() {
		t.Log("Cleaning test user data")
		db.Where("user_id = ?", testUser.ID).Delete(&models.UserHistory{})
		db.Delete(&testUser)
	}()

	// Get 1st game
	history1, puzzle1, err := repo.GetGameRound(ctx, testUser.ID)
	if err != nil {
		t.Fatalf("Failed to get unplayed puzzle: %v", err)
	}
	if history1.Status != models.StatusPlaying || history1.Attempts != 0 {
		t.Errorf("Expected new game status to be 'playing' with 0 aaempts, BUT got status '%s' with %d attempts", history1.Status, history1.Attempts)
	} else {
		t.Logf("New game started. Puzzle ID : %d, Status: %s, Attempts: %d", puzzle1.ID, history1.Status, history1.Attempts)
	}
	// Simulating wrong attempts
	repo.IncrementAttempts(ctx, history1.ID)
	repo.IncrementAttempts(ctx, history1.ID)
	// Verify the database records
	updatedHistory, _, err := repo.GetActiveRoundByID(ctx, testUser.ID, history1.ID)
	if err != nil {
		t.Fatalf("Failed to fetch active round: %v", err)
	}
	if updatedHistory.Attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", updatedHistory.Attempts)
	} else {
		t.Log("Database correctly tracked 2 attempts.")
	}
	// Simulate Correct answer
	err = repo.CompleteRound(ctx, testUser.ID, history1.ID, models.StatusWon)
	if err != nil {
		t.Fatalf("Failed to complete round: %v", err)
	}
	t.Log("Player correctly guessed the hero.")

	// Verify history record
	var updatedUser models.User
	db.First(&updatedUser, testUser.ID)
	if updatedUser.Wins != 1 {
		t.Errorf("Expected wins count to be 1, but got %d", updatedUser.Wins)
	} else {
		t.Logf("User wins count correctly updated to %d", updatedUser.Wins)
	}

	var finalHistory models.UserHistory
	db.First(&finalHistory, history1.ID)
	if finalHistory.Status != models.StatusWon {
		t.Errorf("Expected history status 'won', but got '%s'", finalHistory.Status)
	} else {
		t.Logf("Victory status saved correctly")
	}
}
