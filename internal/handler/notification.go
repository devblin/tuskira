package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/service"
	"github.com/labstack/echo/v4"
)

type NotificationHandler struct {
	svc *service.NotificationService
}

func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{svc: svc}
}

type SendRequest struct {
	Recipient  string        `json:"recipient" validate:"required"`
	Channel    model.Channel `json:"channel" validate:"required"`
	Subject    string        `json:"subject"`
	Body       string        `json:"body"`
	TemplateID   *uint              `json:"template_id"`
	TemplateData model.TemplateData `json:"template_data,omitempty"`
	ScheduleAt   *string            `json:"schedule_at"` // RFC3339
}

func (h *NotificationHandler) Send(c echo.Context) error {
	var req SendRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	n := &model.Notification{
		Recipient:  req.Recipient,
		Channel:    req.Channel,
		Subject:    req.Subject,
		Body:       req.Body,
		TemplateID:   req.TemplateID,
		TemplateData: req.TemplateData,
	}

	if req.ScheduleAt != nil {
		t, err := time.Parse(time.RFC3339, *req.ScheduleAt)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule_at format, use RFC3339"})
		}
		n.ScheduleAt = &t
	}

	if err := h.svc.Send(n); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, n)
}

func (h *NotificationHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	n, err := h.svc.GetByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "notification not found"})
	}

	return c.JSON(http.StatusOK, n)
}

func (h *NotificationHandler) ListByRecipient(c echo.Context) error {
	recipient := c.QueryParam("recipient")
	if recipient == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "recipient query param required"})
	}

	notifications, err := h.svc.ListByRecipient(recipient)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) ListSent(c echo.Context) error {
	notifications, err := h.svc.ListSent()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) ListPending(c echo.Context) error {
	notifications, err := h.svc.ListPending()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) GetPendingScheduled(c echo.Context) error {
	notifications, err := h.svc.GetPendingScheduled()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, notifications)
}

func (h *NotificationHandler) TriggerSend(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	n, err := h.svc.SendByID(uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, n)
}

type UpdateScheduleRequest struct {
	ScheduleAt string `json:"schedule_at" validate:"required"`
}

func (h *NotificationHandler) UpdateSchedule(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	var req UpdateScheduleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	t, err := time.Parse(time.RFC3339, req.ScheduleAt)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid schedule_at format, use RFC3339"})
	}

	n, err := h.svc.UpdateSchedule(uint(id), t)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, n)
}

func (h *NotificationHandler) CancelScheduled(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid id"})
	}

	n, err := h.svc.CancelScheduled(uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, n)
}
