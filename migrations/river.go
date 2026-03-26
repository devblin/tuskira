package migrations

import (
	"context"
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"gorm.io/gorm"
)

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

func RunRiverMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return fmt.Errorf("failed to create river migrator: %w", err)
	}

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, nil)
	if err != nil {
		return fmt.Errorf("failed to run river migrations: %w", err)
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS river_job_id_map (
			external_id TEXT PRIMARY KEY,
			river_job_id BIGINT NOT NULL
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create river_job_id_map table: %w", err)
	}

	return nil
}
