package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/devblin/tuskira/internal/model"
	"github.com/wneessen/go-mail"
)

// EmailProvider sends notifications via SMTP using go-mail.
type EmailProvider struct{}

func NewEmailProvider() *EmailProvider {
	return &EmailProvider{}
}

func (p *EmailProvider) Channel() model.Channel {
	return model.ChannelEmail
}

func (p *EmailProvider) Send(n *model.Notification, rawCfg json.RawMessage) error {
	var cfg model.EmailChannelConfig
	if err := json.Unmarshal(rawCfg, &cfg); err != nil {
		return fmt.Errorf("invalid email channel config: %w", err)
	}

	if cfg.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	port := 587
	if cfg.Port != "" {
		var err error
		port, err = strconv.Atoi(cfg.Port)
		if err != nil {
			return fmt.Errorf("invalid SMTP port: %w", err)
		}
	}

	msg := mail.NewMsg()
	if err := msg.From(cfg.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if err := msg.To(n.Recipient); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	msg.Subject(n.Subject)
	msg.SetBodyString(mail.TypeTextPlain, n.Body)

	opts := []mail.Option{
		mail.WithPort(port),
		mail.WithUsername(cfg.Username),
		mail.WithPassword(cfg.Password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
	}
	if cfg.TLS {
		opts = append(opts, mail.WithSSLPort(false))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", n.Recipient, err)
	}

	log.Printf("[EMAIL] sent to %s | Subject: %s", n.Recipient, n.Subject)
	return nil
}
