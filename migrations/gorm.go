package migrations

import (
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"gorm.io/gorm"
)

// RunGormMigrations auto-migrates all application models.
func RunGormMigrations(db *gorm.DB) error {
	// Backfill user_id on existing rows before AutoMigrate adds NOT NULL constraints.
	if err := backfillUserID(db); err != nil {
		return fmt.Errorf("failed to backfill user_id: %w", err)
	}

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

// backfillUserID adds a nullable user_id column and sets it to the first user's ID
// on tables that predate the user-scoped ownership change.
func backfillUserID(db *gorm.DB) error {
	tables := []string{"notifications", "templates", "channel_configs"}
	for _, table := range tables {
		var hasCol int64
		db.Raw("SELECT COUNT(*) FROM information_schema.columns WHERE table_name = ? AND column_name = 'user_id'", table).Scan(&hasCol)
		if hasCol > 0 {
			continue // column already exists
		}

		// Add as nullable first
		if err := db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN user_id bigint", table)).Error; err != nil {
			return fmt.Errorf("add user_id to %s: %w", table, err)
		}

		// Backfill with the first user's ID
		if err := db.Exec(fmt.Sprintf("UPDATE %s SET user_id = (SELECT id FROM users ORDER BY id LIMIT 1) WHERE user_id IS NULL", table)).Error; err != nil {
			return fmt.Errorf("backfill user_id on %s: %w", table, err)
		}
	}
	return nil
}
