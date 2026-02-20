package db

import (
	"fmt"
	"log"

	"github.com/iamsr/virallens/backend/internal/config"
	"github.com/iamsr/virallens/backend/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewDatabase initializes a new GORM Postgres connection
func NewDatabase(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg.Database.SSLMode)

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	log.Println("Successfully connected to Postgres via GORM")

	// Auto-Migrate domain models
	log.Println("Running AutoMigration...")
	err = db.AutoMigrate(
		&models.User{},
		&models.RefreshToken{},
		&models.Conversation{},
		&models.Group{},
		&models.GroupMember{},
		&models.Message{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to run AutoMigrate: %w", err)
	}
	log.Println("AutoMigration completed.")

	return db, nil
}
