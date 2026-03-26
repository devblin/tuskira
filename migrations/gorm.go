package migrations

import (
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"gorm.io/gorm"
)

// RunGormMigrations auto-migrates all application models.
func RunGormMigrations(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Notification{},
		&model.Template{},
		&model.ChannelConfig{},
	); err != nil {
		return fmt.Errorf("failed to run gorm migrations: %w", err)
	}
	return nil
}
