package provider

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/sse"
)

var ErrClientNotConnected = errors.New("inapp client is not connected")

// InAppProvider delivers notifications to connected SSE clients via the Hub.
// If the client isn't connected, returns ErrClientNotConnected so the notification
// can be saved as pending and replayed when the client reconnects.
type InAppProvider struct {
	hub *sse.Hub
}

func NewInAppProvider(hub *sse.Hub) *InAppProvider {
	return &InAppProvider{hub: hub}
}

func (p *InAppProvider) Channel() model.Channel {
	return model.ChannelInApp
}

func (p *InAppProvider) Send(n *model.Notification, rawCfg json.RawMessage) error {
	var cfg model.InAppChannelConfig
	if err := json.Unmarshal(rawCfg, &cfg); err != nil {
		return fmt.Errorf("failed to parse inapp config: %w", err)
	}

	if cfg.ConnectionID == "" {
		return fmt.Errorf("inapp channel has no connection_id configured")
	}

	msg := &sse.Message{
		NotificationID: n.ID,
		Subject:        n.Subject,
		Body:           n.Body,
		Recipient:      n.Recipient,
	}

	if err := p.hub.Send(cfg.ConnectionID, msg); err != nil {
		return fmt.Errorf("%w: %s", ErrClientNotConnected, err.Error())
	}
	return nil
}
