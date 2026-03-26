package queue

import (
	"context"
	"fmt"

	riverprovider "github.com/devblin/tuskira/pkg/queue/river"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	Type    string
	Payload []byte
}

type HandlerFunc func(ctx context.Context, taskType string, payload []byte) error

// Queue is the interface for async task processing. Implementations handle
// enqueueing tasks, setting worker handlers, and lifecycle management.
type Queue interface {
	Enqueue(ctx context.Context, task Task) error
	SetHandler(handler HandlerFunc)
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Config struct {
	Provider string
	Pool     *pgxpool.Pool
}

func New(cfg Config) (Queue, error) {
	switch cfg.Provider {
	case "river":
		rq, err := riverprovider.New(cfg.Pool)
		if err != nil {
			return nil, fmt.Errorf("failed to create river queue: %w", err)
		}
		return &riverAdapter{inner: rq}, nil
	default:
		return nil, fmt.Errorf("unsupported queue provider: %s", cfg.Provider)
	}
}

type riverAdapter struct {
	inner *riverprovider.Queue
}

func (a *riverAdapter) Enqueue(ctx context.Context, task Task) error {
	return a.inner.Enqueue(ctx, task.Type, task.Payload)
}

func (a *riverAdapter) SetHandler(handler HandlerFunc) { a.inner.SetHandler(handler) }
func (a *riverAdapter) Start(ctx context.Context) error { return a.inner.Start(ctx) }
func (a *riverAdapter) Stop(ctx context.Context) error  { return a.inner.Stop(ctx) }
