package inngest

import (
	"context"
	"log"

	"github.com/inngest/inngestgo"
)

type Queue struct {
	client inngestgo.Client
}

func New(client inngestgo.Client) *Queue {
	return &Queue{client: client}
}

func (q *Queue) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	_, err := q.client.Send(ctx, inngestgo.Event{
		Name: taskType,
		Data: map[string]any{
			"payload": string(payload),
		},
	})
	if err != nil {
		return err
	}
	log.Printf("[INNGEST QUEUE] enqueued task: %s", taskType)
	return nil
}
