package smtp

import (
	"fmt"

	"github.com/devblin/tuskira/pkg/email"
	"github.com/wneessen/go-mail"
)

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	TLS      bool
}

type SMTPSender struct {
	cfg Config
}

func New(cfg Config) *SMTPSender {
	return &SMTPSender{cfg: cfg}
}

func (s *SMTPSender) Send(msg email.Message) error {
	m := mail.NewMsg()
	if err := m.From(msg.From); err != nil {
		return fmt.Errorf("invalid from address: %w", err)
	}
	if err := m.To(msg.To); err != nil {
		return fmt.Errorf("invalid recipient address: %w", err)
	}
	m.Subject(msg.Subject)
	m.SetBodyString(mail.TypeTextPlain, msg.Body)

	opts := []mail.Option{
		mail.WithPort(s.cfg.Port),
		mail.WithUsername(s.cfg.Username),
		mail.WithPassword(s.cfg.Password),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
	}
	if s.cfg.TLS {
		opts = append(opts, mail.WithSSLPort(false))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSOpportunistic))
	}

	client, err := mail.NewClient(s.cfg.Host, opts...)
	if err != nil {
		return fmt.Errorf("failed to create mail client: %w", err)
	}

	if err := client.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
