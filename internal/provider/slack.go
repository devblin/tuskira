package provider

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/devblin/tuskira/internal/model"
	"github.com/slack-go/slack"
)

// SlackProvider sends notifications to Slack channels using the Slack API.
type SlackProvider struct{}

func NewSlackProvider() *SlackProvider {
	return &SlackProvider{}
}

func (p *SlackProvider) Channel() model.Channel {
	return model.ChannelSlack
}

func (p *SlackProvider) Send(n *model.Notification, rawCfg json.RawMessage) error {
	var cfg model.SlackChannelConfig
	if err := json.Unmarshal(rawCfg, &cfg); err != nil {
		return fmt.Errorf("invalid slack channel config: %w", err)
	}

	if cfg.BotToken == "" {
		return fmt.Errorf("slack bot token not configured")
	}

	channel := n.Recipient
	if channel == "" {
		return fmt.Errorf("no slack channel specified in recipient")
	}

	text := n.Body
	if n.Subject != "" {
		text = fmt.Sprintf("*%s*\n%s", n.Subject, n.Body)
	}

	client := slack.New(cfg.BotToken)
	_, _, err := client.PostMessage(channel, slack.MsgOptionText(text, false))
	if err != nil {
		return fmt.Errorf("failed to send slack message to %s: %w", channel, err)
	}

	log.Printf("[SLACK] sent to %s", channel)
	return nil
}
