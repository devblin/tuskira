package queue

import (
	"context"
	"fmt"

	inngestprovider "github.com/devblin/tuskira/pkg/queue/inngest"
	"github.com/inngest/inngestgo"
)

type Task struct {
	Type    string
	Payload []byte
}

type Queue interface {
	Enqueue(ctx context.Context, task Task) error
}

type Config struct {
	Provider string
	EventKey string
}

func New(cfg Config) (Queue, error) {
	switch cfg.Provider {
	case "inngest":
		client, err := inngestgo.NewClient(inngestgo.ClientOpts{EventKey: &cfg.EventKey})
		if err != nil {
			return nil, fmt.Errorf("failed to create inngest client: %w", err)
		}
		return &inngestAdapter{inner: inngestprovider.New(client)}, nil
	default:
		return nil, fmt.Errorf("unsupported queue provider: %s", cfg.Provider)
	}
}

type inngestAdapter struct {
	inner *inngestprovider.Queue
}

func (a *inngestAdapter) Enqueue(ctx context.Context, task Task) error {
	return a.inner.Enqueue(ctx, task.Type, task.Payload)
}
