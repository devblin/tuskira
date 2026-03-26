package database

import (
	"fmt"
	"log"

	"github.com/devblin/tuskira/internal/config"
	"github.com/devblin/tuskira/internal/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init(cfg *config.Config) *gorm.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.Notification{},
		&model.Template{},
	); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}

	return db
}
