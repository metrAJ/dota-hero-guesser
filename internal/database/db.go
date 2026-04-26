package database

import (
	"doproj/internal/models"
	"fmt"
	"log"

	"doproj/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) *gorm.DB {
	// Connecting to DB, logging Errors only
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s", cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort, cfg.DBSSLMode, cfg.DBTimeZone)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	err = db.AutoMigrate(&models.Item{}, &models.Hero{}, &models.Puzzle{}, &models.User{}, &models.UserHistory{}, &models.MultiplayerMatch{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}
	return db
}
