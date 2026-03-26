package river

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type GenericJobArgs struct {
	TaskType string `json:"task_type"`
	Payload  []byte `json:"payload"`
}

func (GenericJobArgs) Kind() string { return "generic_task" }

type GenericJobWorker struct {
	river.WorkerDefaults[GenericJobArgs]
	Handler func(ctx context.Context, taskType string, payload []byte) error
}

func (w *GenericJobWorker) Work(ctx context.Context, job *river.Job[GenericJobArgs]) error {
	if w.Handler == nil {
		log.Printf("[RIVER QUEUE] no handler set, skipping task: %s", job.Args.TaskType)
		return nil
	}
	return w.Handler(ctx, job.Args.TaskType, job.Args.Payload)
}

// Queue is the River-backed implementation of the queue interface.
// It wraps a River client and routes all tasks through a single GenericJobWorker.
type Queue struct {
	client *river.Client[pgx.Tx]
	worker *GenericJobWorker
}

func New(pool *pgxpool.Pool) (*Queue, error) {
	worker := &GenericJobWorker{}
	workers := river.NewWorkers()
	river.AddWorker(workers, worker)

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Workers: workers,
		Queues: map[string]river.QueueConfig{
			"tasks": {MaxWorkers: 100},
		},
	})
	if err != nil {
		return nil, err
	}
	return &Queue{client: client, worker: worker}, nil
}

func (q *Queue) SetHandler(handler func(ctx context.Context, taskType string, payload []byte) error) {
	q.worker.Handler = handler
}

func (q *Queue) Start(ctx context.Context) error {
	return q.client.Start(ctx)
}

func (q *Queue) Stop(ctx context.Context) error {
	return q.client.Stop(ctx)
}

func (q *Queue) Enqueue(ctx context.Context, taskType string, payload []byte) error {
	_, err := q.client.Insert(ctx, GenericJobArgs{
		TaskType: taskType,
		Payload:  payload,
	}, &river.InsertOpts{Queue: "tasks"})
	if err != nil {
		return err
	}
	log.Printf("[RIVER QUEUE] enqueued task: %s", taskType)
	return nil
}
