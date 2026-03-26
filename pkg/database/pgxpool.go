package database

import (
	"context"
	"fmt"

	"github.com/devblin/tuskira/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPgxPool creates a pgx connection pool, used by River queue/scheduler.
func NewPgxPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName,
	)
	return pgxpool.New(ctx, dsn)
}
