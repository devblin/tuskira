package provider

import (
	"encoding/json"
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

func (p *InAppProvider) Send(n *model.Notification, _ json.RawMessage) error {
	log.Printf("[INAPP] To: %s | Body: %s", n.Recipient, n.Body)
	return nil
}
