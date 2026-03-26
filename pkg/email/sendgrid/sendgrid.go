package sendgrid

import (
	"fmt"
	"net/http"

	"github.com/devblin/tuskira/pkg/email"
	"github.com/sendgrid/sendgrid-go"
	sgmail "github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Config struct {
	APIKey string
}

type SendGridSender struct {
	cfg Config
}

func New(cfg Config) *SendGridSender {
	return &SendGridSender{cfg: cfg}
}

func (s *SendGridSender) Send(msg email.Message) error {
	from := sgmail.NewEmail("", msg.From)
	to := sgmail.NewEmail("", msg.To)
	message := sgmail.NewSingleEmail(from, msg.Subject, to, msg.Body, msg.Body)

	client := sendgrid.NewSendClient(s.cfg.APIKey)
	resp, err := client.Send(message)
	if err != nil {
		return fmt.Errorf("sendgrid request failed: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("sendgrid returned status %d: %s", resp.StatusCode, resp.Body)
	}

	return nil
}
