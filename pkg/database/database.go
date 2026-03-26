package database

import (
	"fmt"

	"github.com/devblin/tuskira/internal/config"
	"github.com/devblin/tuskira/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Init opens a GORM database connection using config credentials.
func Init(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// SeedDefaultUser creates a default test user if one doesn't already exist.
func SeedDefaultUser(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Where("email = ?", "test@email.com").Count(&count)
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("12345678"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash default password: %w", err)
	}
	return db.Create(&model.User{
		Email:    "test@email.com",
		Password: string(hash),
	}).Error
}
