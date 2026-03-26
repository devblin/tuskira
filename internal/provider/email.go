package provider

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/devblin/tuskira/internal/model"
	pkgemail "github.com/devblin/tuskira/pkg/email"
	"github.com/devblin/tuskira/pkg/email/sendgrid"
	"github.com/devblin/tuskira/pkg/email/smtp"
)

type EmailProvider struct{}

func NewEmailProvider() *EmailProvider {
	return &EmailProvider{}
}

func (p *EmailProvider) Channel() model.Channel {
	return model.ChannelEmail
}

func (p *EmailProvider) Send(n *model.Notification, rawCfg json.RawMessage) error {
	var provCfg model.EmailProviderConfig

	// Use provider config from notification if available (selected by user at send time)
	if len(n.ProviderConfig) > 0 {
		if err := json.Unmarshal(n.ProviderConfig, &provCfg); err != nil {
			return fmt.Errorf("invalid provider config on notification: %w", err)
		}
	} else {
		// Fall back to first provider in channel config
		var cfg model.EmailChannelConfig
		if err := json.Unmarshal(rawCfg, &cfg); err != nil {
			return fmt.Errorf("invalid email channel config: %w", err)
		}
		if len(cfg.Providers) == 0 {
			return fmt.Errorf("no email providers configured")
		}
		provCfg = cfg.Providers[0]
	}

	sender, err := p.resolveSender(provCfg)
	if err != nil {
		return err
	}

	msg := pkgemail.Message{
		From:    provCfg.From,
		To:      n.Recipient,
		Subject: n.Subject,
		Body:    n.Body,
	}

	if err := sender.Send(msg); err != nil {
		return fmt.Errorf("failed to send email to %s: %w", n.Recipient, err)
	}

	provider := provCfg.Provider
	if provider == "" {
		provider = "smtp"
	}
	log.Printf("[EMAIL:%s] sent to %s | Subject: %s", provider, n.Recipient, n.Subject)
	return nil
}

func (p *EmailProvider) resolveSender(cfg model.EmailProviderConfig) (pkgemail.Sender, error) {
	provider := cfg.Provider
	if provider == "" {
		provider = "smtp"
	}

	switch provider {
	case "smtp":
		if cfg.Host == "" {
			return nil, fmt.Errorf("SMTP host not configured")
		}
		port := 587
		if cfg.Port != "" {
			var err error
			port, err = strconv.Atoi(cfg.Port)
			if err != nil {
				return nil, fmt.Errorf("invalid SMTP port: %w", err)
			}
		}
		return smtp.New(smtp.Config{
			Host:     cfg.Host,
			Port:     port,
			Username: cfg.Username,
			Password: cfg.Password,
			TLS:      cfg.TLS,
		}), nil
	case "sendgrid":
		if cfg.APIKey == "" {
			return nil, fmt.Errorf("SendGrid API key not configured")
		}
		return sendgrid.New(sendgrid.Config{
			APIKey: cfg.APIKey,
		}), nil
	default:
		return nil, fmt.Errorf("unsupported email provider: %s", provider)
	}
}
