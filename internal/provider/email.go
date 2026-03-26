package provider

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/smtp"

	"github.com/devblin/tuskira/internal/model"
)

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

	addr := net.JoinHostPort(cfg.Host, cfg.Port)
	if cfg.Port == "" {
		addr = net.JoinHostPort(cfg.Host, "587")
	}

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		cfg.From, n.Recipient, n.Subject, n.Body)

	auth := smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)

	var err error
	if cfg.TLS {
		err = sendWithTLS(addr, cfg.Host, auth, cfg.From, n.Recipient, []byte(msg))
	} else {
		err = smtp.SendMail(addr, auth, cfg.From, []string{n.Recipient}, []byte(msg))
	}

	if err != nil {
		return fmt.Errorf("failed to send email to %s: %w", n.Recipient, err)
	}

	log.Printf("[EMAIL] sent to %s | Subject: %s", n.Recipient, n.Subject)
	return nil
}

func sendWithTLS(addr, host string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsConfig := &tls.Config{ServerName: host}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS dial failed: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("SMTP client creation failed: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("SMTP RCPT TO failed: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("SMTP write failed: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("SMTP data close failed: %w", err)
	}

	return client.Quit()
}
