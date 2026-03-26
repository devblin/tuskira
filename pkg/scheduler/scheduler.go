package scheduler

import (
	"context"
	"fmt"
	"time"

	riverprovider "github.com/devblin/tuskira/pkg/scheduler/river"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Job struct {
	ID      string
	Payload []byte
	RunAt   time.Time
}

type HandlerFunc func(ctx context.Context, externalID string, payload []byte) error

// Scheduler is the interface for delayed job execution. Jobs are scheduled
// at a specific time and can be cancelled or rescheduled.
type Scheduler interface {
	Schedule(ctx context.Context, job Job) error
	Cancel(ctx context.Context, jobID string) error
	Reschedule(ctx context.Context, jobID string, newTime time.Time) error
	SetHandler(handler HandlerFunc)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Config struct {
	Provider string
	Pool     *pgxpool.Pool
}

func New(cfg Config) (Scheduler, error) {
	switch cfg.Provider {
	case "river":
		rs, err := riverprovider.New(cfg.Pool)
		if err != nil {
			return nil, fmt.Errorf("failed to create river scheduler: %w", err)
		}
		return &riverAdapter{inner: rs}, nil
	default:
		return nil, fmt.Errorf("unsupported scheduler provider: %s", cfg.Provider)
	}
}

type riverAdapter struct {
	inner *riverprovider.Scheduler
}

func (a *riverAdapter) Schedule(ctx context.Context, job Job) error {
	return a.inner.Schedule(ctx, job.ID, job.Payload, job.RunAt)
}

func (a *riverAdapter) Cancel(ctx context.Context, jobID string) error {
	return a.inner.Cancel(ctx, jobID)
}

func (a *riverAdapter) Reschedule(ctx context.Context, jobID string, newTime time.Time) error {
	return a.inner.Reschedule(ctx, jobID, newTime)
}

func (a *riverAdapter) SetHandler(handler HandlerFunc) { a.inner.SetHandler(handler) }
func (a *riverAdapter) Start(ctx context.Context) error { return a.inner.Start(ctx) }
func (a *riverAdapter) Stop(ctx context.Context) error  { return a.inner.Stop(ctx) }
