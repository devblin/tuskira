package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/devblin/tuskira/internal/service"
	"github.com/devblin/tuskira/internal/sse"
	"github.com/labstack/echo/v4"
)

type SSEHandler struct {
	hub              *sse.Hub
	channelConfigSvc *service.ChannelConfigService
	notifRepo        *repository.NotificationRepository
}

func NewSSEHandler(hub *sse.Hub, channelConfigSvc *service.ChannelConfigService, notifRepo *repository.NotificationRepository) *SSEHandler {
	return &SSEHandler{hub: hub, channelConfigSvc: channelConfigSvc, notifRepo: notifRepo}
}

// Stream opens an SSE connection. Validates the connection_id against the inapp channel config,
// replays any pending notifications, then streams new messages in real time.
func (h *SSEHandler) Stream(c echo.Context) error {
	connectionID := c.QueryParam("connection_id")
	if connectionID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "connection_id query param required"})
	}

	userID := c.Get("user_id").(uint)
	cfg, err := h.channelConfigSvc.GetByChannel(model.ChannelInApp, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "inapp channel is not configured"})
	}
	if !cfg.Enabled {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "inapp channel is disabled"})
	}

	var inappCfg model.InAppChannelConfig
	if err := json.Unmarshal(cfg.Config, &inappCfg); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to parse inapp config"})
	}
	if inappCfg.ConnectionID != connectionID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid connection_id"})
	}

	w := c.Response()
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	w.Flush()

	client := h.hub.Register(connectionID)
	defer h.hub.Unregister(connectionID)

	// Replay pending/failed in-app notifications
	pending, err := h.notifRepo.FindPendingByRecipientAndChannel(connectionID, model.ChannelInApp)
	if err != nil {
		log.Printf("[SSE] failed to load pending notifications for %s: %v", connectionID, err)
	} else {
		for i := range pending {
			msg := &sse.Message{
				NotificationID: pending[i].ID,
				Subject:        pending[i].Subject,
				Body:           pending[i].Body,
				Recipient:      pending[i].Recipient,
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
			}
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[SSE] failed to marshal notification %d: %v", pending[i].ID, err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.Flush()

			now := time.Now()
			pending[i].Status = model.StatusSent
			pending[i].SentAt = &now
			if err := h.notifRepo.Save(&pending[i]); err != nil {
				log.Printf("[SSE] failed to update notification %d status: %v", pending[i].ID, err)
			}
		}
	}

	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	ctx := c.Request().Context()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-client.Done:
			return nil
		case msg := <-client.Messages:
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("[SSE] failed to marshal message: %v", err)
				continue
			}
			if _, err := fmt.Fprintf(w, "data: %s\n\n", data); err != nil {
				return nil
			}
			w.Flush()
		case <-keepalive.C:
			if _, err := fmt.Fprintf(w, ": keepalive\n\n"); err != nil {
				return nil
			}
			w.Flush()
		}
	}
}
