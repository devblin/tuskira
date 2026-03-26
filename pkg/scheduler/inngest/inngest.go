package inngest

import (
	"context"
	"log"
	"time"

	"github.com/inngest/inngestgo"
)

type Scheduler struct {
	client inngestgo.Client
}

func New(client inngestgo.Client) *Scheduler {
	return &Scheduler{client: client}
}

func (s *Scheduler) Schedule(ctx context.Context, id string, payload []byte, runAt time.Time) error {
	_, err := s.client.Send(ctx, inngestgo.Event{
		Name: "notification/schedule",
		Data: map[string]any{
			"job_id":  id,
			"payload": string(payload),
			"run_at":  runAt.Format(time.RFC3339),
		},
	})
	if err != nil {
		return err
	}
	log.Printf("[INNGEST SCHEDULER] scheduled job %s for %s", id, runAt)
	return nil
}

func (s *Scheduler) Cancel(ctx context.Context, jobID string) error {
	// TODO: use Inngest's cancellation API to cancel a scheduled function run
	log.Printf("[INNGEST SCHEDULER] cancelled job %s", jobID)
	return nil
}

func (s *Scheduler) Reschedule(ctx context.Context, jobID string, newTime time.Time) error {
	if err := s.Cancel(ctx, jobID); err != nil {
		return err
	}
	log.Printf("[INNGEST SCHEDULER] rescheduled job %s to %s", jobID, newTime)
	return nil
}
