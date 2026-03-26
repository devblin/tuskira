package provider

import (
	"log"

	"github.com/devblin/tuskira/internal/model"
)

type InAppProvider struct{}

func NewInAppProvider() *InAppProvider {
	return &InAppProvider{}
}

func (p *InAppProvider) Channel() model.Channel {
	return model.ChannelInApp
}

func (p *InAppProvider) Send(n *model.Notification) error {
	// TODO: persist in-app notification and push via WebSocket/SSE
	log.Printf("[INAPP] To: %s | Body: %s", n.Recipient, n.Body)
	return nil
}
