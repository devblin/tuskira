package provider

import (
	"log"

	"github.com/devblin/tuskira/internal/model"
)

type SlackProvider struct{}

func NewSlackProvider() *SlackProvider {
	return &SlackProvider{}
}

func (p *SlackProvider) Channel() model.Channel {
	return model.ChannelSlack
}

func (p *SlackProvider) Send(n *model.Notification) error {
	// TODO: integrate with Slack API (e.g. slack-go/slack)
	log.Printf("[SLACK] To: %s | Body: %s", n.Recipient, n.Body)
	return nil
}
