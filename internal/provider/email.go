package provider

import (
	"log"

	"github.com/devblin/tuskira/internal/model"
)

type EmailProvider struct{}

func NewEmailProvider() *EmailProvider {
	return &EmailProvider{}
}

func (p *EmailProvider) Channel() model.Channel {
	return model.ChannelEmail
}

func (p *EmailProvider) Send(n *model.Notification) error {
	// TODO: integrate with an email service (e.g. SendGrid, SES, SMTP)
	log.Printf("[EMAIL] To: %s | Subject: %s | Body: %s", n.Recipient, n.Subject, n.Body)
	return nil
}
