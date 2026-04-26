package repository

import (
	"context"
	"doproj/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormRepo struct {
	db *gorm.DB
}

func NewGormRepo(db *gorm.DB) *gormRepo {
	return &gormRepo{db: db}
}

func (r *gormRepo) SaveHero(hero *models.Hero) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(hero).Error
}

func (r *gormRepo) SaveItem(item *models.Item) error {
	return r.db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(item).Error
}

func (r *gormRepo) SavePuzzle(puzzle *models.Puzzle) error {
	return r.db.Create(puzzle).Error
}

func (r *gormRepo) GetUnplayedPuzzle(ctx context.Context, userID uint) (*models.Puzzle, error) {
	var puzzle models.Puzzle
	err := r.db.WithContext(ctx).Preload("Hero").Where("id NOT IN (?)", r.db.Table("user_histories").Select("puzzle_id").Where("user_id = ?", userID)).Order("RANDOM()").First(&puzzle).Error
	return &puzzle, err
}

func (r *gormRepo) GetItemsByIDs(ctx context.Context, itemIDs []int64) ([]models.Item, error) {
	result := []models.Item{}
	if len(itemIDs) == 0 {
		return result, nil
	}
	var dbItems []models.Item
	err := r.db.WithContext(ctx).Where("id IN ?", itemIDs).Find(&dbItems).Error
	if err != nil {
		return nil, err
	}
	itemMap := make(map[int64]models.Item)
	for _, item := range dbItems {
		itemMap[int64(item.ID)] = item
	}
	for _, id := range itemIDs {
		if item, exists := itemMap[id]; exists {
			result = append(result, item)
		}
	}
	return result, nil
}

func (r *gormRepo) GetHeroByID(ctx context.Context, heroID uint) (*models.Hero, error) {
	var hero models.Hero
	err := r.db.WithContext(ctx).First(&hero, heroID).Error
	return &hero, err
}

func (r *gormRepo) GetGameRound(ctx context.Context, userID uint) (*models.UserHistory, *models.Puzzle, error) {
	var history models.UserHistory
	// Checking for active round
	err := r.db.WithContext(ctx).Where("user_id = ? AND status = ?", userID, models.StatusPlaying).First(&history).Error
	if err == nil {
		// Returning active puzzle
		var puzzle models.Puzzle
		err = r.db.WithContext(ctx).First(&puzzle, history.PuzzleID).Error
		return &history, &puzzle, err
	}
	// If no active, creating new round
	puzzle, err := r.GetUnplayedPuzzle(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	history = models.UserHistory{
		UserID:   userID,
		PuzzleID: puzzle.ID,
		Status:   models.StatusPlaying,
		Attempts: 0,
	}
	err = r.db.WithContext(ctx).Create(&history).Error
	return &history, puzzle, err
}

func (r *gormRepo) GetActiveRoundByID(ctx context.Context, userID uint, roundID uint) (*models.UserHistory, *models.Puzzle, error) {
	var history models.UserHistory
	var puzzle models.Puzzle

	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", roundID, userID).First(&history).Error
	if err != nil {
		return nil, nil, err
	}

	err = r.db.WithContext(ctx).First(&puzzle, history.PuzzleID).Error
	return &history, &puzzle, err
}

func (r *gormRepo) IncrementAttempts(ctx context.Context, roundID uint) error {
	return r.db.WithContext(ctx).Model(&models.UserHistory{}).
		Where("id = ?", roundID).
		Update("attempts", gorm.Expr("attempts + 1")).Error
}

func (r *gormRepo) CompleteRound(ctx context.Context, userID uint, roundID uint, finalStatus string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Model(&models.UserHistory{}).Where("id = ?", roundID).
			Update("status", finalStatus).Error; err != nil {
			return err
		}
		statField := ""
		switch finalStatus {
		case models.StatusWon:
			statField = "wins"
		case models.StatusLost:
			statField = "losses"
		}
		if statField != "" {
			if err := tx.Model(&models.User{}).Where("id = ?", userID).
				Update(statField, gorm.Expr(statField+" + 1")).Error; err != nil {
				return err
			}
		}
		return nil
	})

}

func (r *gormRepo) IncrementWins(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("wins", gorm.Expr("wins + 1")).Error
}

func (r *gormRepo) IncrementLosses(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("losses", gorm.Expr("losses + 1")).Error
}

func (r *gormRepo) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *gormRepo) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	return &user, err
}

func (r *gormRepo) GetUserByID(ctx context.Context, userID uint) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error
	return &user, err
}

func (r *gormRepo) GetAllHeroes() ([]models.Hero, error) {
	var heroes []models.Hero
	err := r.db.Find(&heroes).Error
	return heroes, err
}

func (r *gormRepo) GetTopPlayers(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Where("wins > 0").Order("wins DESC").Limit(10).Find(&users).Error
	return users, err
}

func (r *gormRepo) GetProcessedMatchIDs(ctx context.Context) (map[uint64]bool, error) {
	var matchIDs []uint64

	// Grab only unique match_ids from the puzzles table
	err := r.db.WithContext(ctx).
		Model(&models.Puzzle{}).
		Distinct("match_id").
		Pluck("match_id", &matchIDs).Error

	if err != nil {
		return nil, err
	}

	// Convert the slice into a map for instant lookups later
	processedMap := make(map[uint64]bool, len(matchIDs))
	for _, id := range matchIDs {
		processedMap[id] = true
	}

	return processedMap, nil
}

func (r *gormRepo) FindUnplayedMutualPuzzle(ctx context.Context, p1ID, p2ID uint) (*models.Puzzle, error) {
	var puzzle models.Puzzle
	query := `
		SELECT p.* FROM puzzles p
		WHERE NOT EXISTS (
			SELECT 1 FROM user_histories h
			WHERE h.puzzle_id = p.id AND h.user_id IN (?, ?)
		)
		AND NOT EXISTS (
			SELECT 1 FROM multiplayer_matches m 
			WHERE m.puzzle_id = p.id AND (
				m.player1_id IN (?, ?) OR m.player2_id IN (?, ?)
			)
		)	
		ORDER BY RANDOM() 
		LIMIT 1;
	`
	err := r.db.WithContext(ctx).Raw(query, p1ID, p2ID, p1ID, p2ID, p1ID, p2ID).Scan(&puzzle).Error
	if err != nil {
		return nil, err
	}
	if puzzle.ID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &puzzle, nil
}

func (r *gormRepo) CreateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error {
	return r.db.WithContext(ctx).Create(match).Error
}

func (r *gormRepo) UpdateMultiplayerMatch(ctx context.Context, match *models.MultiplayerMatch) error {
	return r.db.WithContext(ctx).Save(match).Error
}
