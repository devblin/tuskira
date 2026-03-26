package scheduler

import (
	"context"
	"fmt"
	"time"

	inngestprovider "github.com/devblin/tuskira/pkg/scheduler/inngest"
	"github.com/inngest/inngestgo"
)

type Job struct {
	ID      string
	Payload []byte
	RunAt   time.Time
}

type Scheduler interface {
	Schedule(ctx context.Context, job Job) error
	Cancel(ctx context.Context, jobID string) error
	Reschedule(ctx context.Context, jobID string, newTime time.Time) error
}

type Config struct {
	Provider string
	EventKey string
	AppID    string
}

func New(cfg Config) (Scheduler, error) {
	switch cfg.Provider {
	case "inngest":
		client, err := inngestgo.NewClient(inngestgo.ClientOpts{AppID: cfg.AppID, EventKey: &cfg.EventKey})
		if err != nil {
			return nil, fmt.Errorf("failed to create inngest client: %w", err)
		}
		return &inngestAdapter{inner: inngestprovider.New(client)}, nil
	default:
		return nil, fmt.Errorf("unsupported scheduler provider: %s", cfg.Provider)
	}
}

type inngestAdapter struct {
	inner *inngestprovider.Scheduler
}

func (a *inngestAdapter) Schedule(ctx context.Context, job Job) error {
	return a.inner.Schedule(ctx, job.ID, job.Payload, job.RunAt)
}

func (a *inngestAdapter) Cancel(ctx context.Context, jobID string) error {
	return a.inner.Cancel(ctx, jobID)
}

func (a *inngestAdapter) Reschedule(ctx context.Context, jobID string, newTime time.Time) error {
	return a.inner.Reschedule(ctx, jobID, newTime)
}
