package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"
	"time"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/provider"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/devblin/tuskira/pkg/queue"
	"github.com/devblin/tuskira/pkg/scheduler"
)

type NotificationService struct {
	repo      *repository.NotificationRepository
	registry  *provider.Registry
	queue     queue.Queue
	scheduler scheduler.Scheduler
}

func NewNotificationService(
	repo *repository.NotificationRepository,
	registry *provider.Registry,
	q queue.Queue,
	s scheduler.Scheduler,
) *NotificationService {
	return &NotificationService{repo: repo, registry: registry, queue: q, scheduler: s}
}

func (s *NotificationService) Send(n *model.Notification) error {
	if n.TemplateID != nil {
		tmpl, err := s.repo.FindTemplateByID(*n.TemplateID)
		if err != nil {
			return fmt.Errorf("template not found: %w", err)
		}
		data := map[string]string(n.TemplateData)
		if n.Subject == "" {
			rendered, err := renderTemplate(tmpl.Subject, data)
			if err != nil {
				return fmt.Errorf("failed to render subject: %w", err)
			}
			n.Subject = rendered
		}
		if n.Body == "" {
			rendered, err := renderTemplate(tmpl.Body, data)
			if err != nil {
				return fmt.Errorf("failed to render body: %w", err)
			}
			n.Body = rendered
		}
	}

	if n.ScheduleAt != nil && n.ScheduleAt.After(time.Now()) {
		n.Status = model.StatusScheduled
		if err := s.repo.Create(n); err != nil {
			return err
		}
		payload, err := json.Marshal(map[string]uint{"notification_id": n.ID})
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		return s.scheduler.Schedule(context.Background(), scheduler.Job{
			ID:      fmt.Sprintf("notification:%d", n.ID),
			Payload: payload,
			RunAt:   *n.ScheduleAt,
		})
	}

	p, ok := s.registry.Get(n.Channel)
	if !ok {
		return fmt.Errorf("no provider registered for channel: %s", n.Channel)
	}

	if err := p.Send(n); err != nil {
		n.Status = model.StatusFailed
		s.repo.Create(n)
		return fmt.Errorf("failed to send notification: %w", err)
	}

	now := time.Now()
	n.Status = model.StatusSent
	n.SentAt = &now
	return s.repo.Create(n)
}

func (s *NotificationService) GetByID(id uint) (*model.Notification, error) {
	return s.repo.FindByID(id)
}

func (s *NotificationService) ListByRecipient(recipient string) ([]model.Notification, error) {
	return s.repo.FindByRecipient(recipient)
}

func (s *NotificationService) GetPendingScheduled() ([]model.Notification, error) {
	return s.repo.FindPendingScheduled()
}

func (s *NotificationService) SendByID(id uint) (*model.Notification, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	if n.Status != model.StatusScheduled {
		return nil, fmt.Errorf("notification %d is not in scheduled status (current: %s)", n.ID, n.Status)
	}

	p, ok := s.registry.Get(n.Channel)
	if !ok {
		return nil, fmt.Errorf("no provider registered for channel: %s", n.Channel)
	}

	if err := p.Send(n); err != nil {
		n.Status = model.StatusFailed
		s.repo.Save(n)
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	now := time.Now()
	n.Status = model.StatusSent
	n.SentAt = &now
	if err := s.repo.Save(n); err != nil {
		return nil, fmt.Errorf("failed to update notification: %w", err)
	}
	return n, nil
}

func (s *NotificationService) UpdateSchedule(id uint, newTime time.Time) (*model.Notification, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	if n.Status != model.StatusScheduled {
		return nil, fmt.Errorf("notification %d is not in scheduled status (current: %s)", n.ID, n.Status)
	}

	jobID := fmt.Sprintf("notification:%d", n.ID)
	if err := s.scheduler.Reschedule(context.Background(), jobID, newTime); err != nil {
		return nil, fmt.Errorf("failed to reschedule: %w", err)
	}

	n.ScheduleAt = &newTime
	if err := s.repo.Save(n); err != nil {
		return nil, fmt.Errorf("failed to update schedule: %w", err)
	}
	return n, nil
}

func (s *NotificationService) CancelScheduled(id uint) (*model.Notification, error) {
	n, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification not found: %w", err)
	}

	if n.Status != model.StatusScheduled {
		return nil, fmt.Errorf("notification %d is not in scheduled status (current: %s)", n.ID, n.Status)
	}

	jobID := fmt.Sprintf("notification:%d", n.ID)
	if err := s.scheduler.Cancel(context.Background(), jobID); err != nil {
		return nil, fmt.Errorf("failed to cancel scheduled job: %w", err)
	}

	n.Status = model.StatusCancelled
	if err := s.repo.Save(n); err != nil {
		return nil, fmt.Errorf("failed to cancel notification: %w", err)
	}
	return n, nil
}

func renderTemplate(tmplStr string, data map[string]string) (string, error) {
	t, err := template.New("").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("invalid template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}
	return buf.String(), nil
}
