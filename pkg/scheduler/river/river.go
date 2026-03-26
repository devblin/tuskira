package river

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type ScheduledJobArgs struct {
	ExternalID string `json:"external_id"`
	Payload    []byte `json:"payload"`
}

func (ScheduledJobArgs) Kind() string { return "scheduled_job" }

type ScheduledJobWorker struct {
	river.WorkerDefaults[ScheduledJobArgs]
	Handler func(ctx context.Context, externalID string, payload []byte) error
}

func (w *ScheduledJobWorker) Work(ctx context.Context, job *river.Job[ScheduledJobArgs]) error {
	if w.Handler == nil {
		log.Printf("[RIVER SCHEDULER] no handler set, skipping job: %s", job.Args.ExternalID)
		return nil
	}
	return w.Handler(ctx, job.Args.ExternalID, job.Args.Payload)
}

type Scheduler struct {
	client *river.Client[pgx.Tx]
	pool   *pgxpool.Pool
	worker *ScheduledJobWorker
}

func New(pool *pgxpool.Pool) (*Scheduler, error) {
	worker := &ScheduledJobWorker{}
	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
	})
	if err != nil {
		return nil, err
	}
	return &Scheduler{client: client, pool: pool, worker: worker}, nil
}

func (s *Scheduler) SetHandler(handler func(ctx context.Context, externalID string, payload []byte) error) {
	s.worker.Handler = handler
}

func (s *Scheduler) Start(ctx context.Context) error {
	return s.client.Start(ctx)
}

func (s *Scheduler) Stop(ctx context.Context) error {
	return s.client.Stop(ctx)
}

func (s *Scheduler) Schedule(ctx context.Context, id string, payload []byte, runAt time.Time) error {
	res, err := s.client.Insert(ctx, ScheduledJobArgs{
		ExternalID: id,
		Payload:    payload,
	}, &river.InsertOpts{
		ScheduledAt: runAt,
	})
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`INSERT INTO river_job_id_map (external_id, river_job_id) VALUES ($1, $2)
		 ON CONFLICT (external_id) DO UPDATE SET river_job_id = $2`,
		id, res.Job.ID)
	if err != nil {
		return fmt.Errorf("failed to store job ID mapping: %w", err)
	}

	log.Printf("[RIVER SCHEDULER] scheduled job %s (river_id=%d) for %s", id, res.Job.ID, runAt)
	return nil
}

func (s *Scheduler) Cancel(ctx context.Context, jobID string) error {
	riverJobID, err := s.lookupRiverID(ctx, jobID)
	if err != nil {
		return err
	}

	_, err = s.client.JobCancel(ctx, riverJobID)
	if err != nil {
		return fmt.Errorf("failed to cancel river job %d: %w", riverJobID, err)
	}

	_, _ = s.pool.Exec(ctx, `DELETE FROM river_job_id_map WHERE external_id = $1`, jobID)

	log.Printf("[RIVER SCHEDULER] cancelled job %s (river_id=%d)", jobID, riverJobID)
	return nil
}

func (s *Scheduler) Reschedule(ctx context.Context, jobID string, newTime time.Time) error {
	riverJobID, err := s.lookupRiverID(ctx, jobID)
	if err != nil {
		return err
	}

	job, err := s.client.JobGet(ctx, riverJobID)
	if err != nil {
		return fmt.Errorf("failed to get river job %d: %w", riverJobID, err)
	}

	_, err = s.client.JobCancel(ctx, riverJobID)
	if err != nil {
		return fmt.Errorf("failed to cancel river job for reschedule: %w", err)
	}

	var args ScheduledJobArgs
	if err := json.Unmarshal(job.EncodedArgs, &args); err != nil {
		return fmt.Errorf("failed to decode job args: %w", err)
	}

	res, err := s.client.Insert(ctx, ScheduledJobArgs{
		ExternalID: jobID,
		Payload:    args.Payload,
	}, &river.InsertOpts{
		ScheduledAt: newTime,
	})
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx,
		`UPDATE river_job_id_map SET river_job_id = $1 WHERE external_id = $2`,
		res.Job.ID, jobID)
	if err != nil {
		return fmt.Errorf("failed to update job ID mapping: %w", err)
	}

	log.Printf("[RIVER SCHEDULER] rescheduled job %s (new river_id=%d) to %s", jobID, res.Job.ID, newTime)
	return nil
}

func (s *Scheduler) lookupRiverID(ctx context.Context, externalID string) (int64, error) {
	var riverJobID int64
	err := s.pool.QueryRow(ctx,
		`SELECT river_job_id FROM river_job_id_map WHERE external_id = $1`,
		externalID).Scan(&riverJobID)
	if err != nil {
		return 0, fmt.Errorf("no river job found for external ID %s: %w", externalID, err)
	}
	return riverJobID, nil
}
