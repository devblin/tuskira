package service

import (
	"encoding/json"
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
)

type ChannelConfigService struct {
	repo *repository.ChannelConfigRepository
}

func NewChannelConfigService(repo *repository.ChannelConfigRepository) *ChannelConfigService {
	return &ChannelConfigService{repo: repo}
}

func (s *ChannelConfigService) Upsert(cfg *model.ChannelConfig) error {
	if err := s.validateConfig(cfg.Channel, cfg.Config); err != nil {
		return fmt.Errorf("invalid config for channel %s: %w", cfg.Channel, err)
	}
	return s.repo.Upsert(cfg)
}

func (s *ChannelConfigService) GetByChannel(channel model.Channel) (*model.ChannelConfig, error) {
	return s.repo.FindByChannel(channel)
}

func (s *ChannelConfigService) List() ([]model.ChannelConfig, error) {
	return s.repo.FindAll()
}

func (s *ChannelConfigService) Delete(channel model.Channel) error {
	return s.repo.Delete(channel)
}

func (s *ChannelConfigService) validateConfig(channel model.Channel, data model.ChannelConfigData) error {
	switch channel {
	case model.ChannelEmail:
		var cfg model.EmailChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("invalid email config: %w", err)
		}
		if cfg.Host == "" {
			return fmt.Errorf("host is required")
		}
		if cfg.From == "" {
			return fmt.Errorf("from is required")
		}
	case model.ChannelSlack:
		var cfg model.SlackChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return fmt.Errorf("invalid slack config: %w", err)
		}
		if cfg.BotToken == "" {
			return fmt.Errorf("bot_token is required")
		}
	case model.ChannelInApp:
		// no config needed
	default:
		return fmt.Errorf("unknown channel: %s", channel)
	}
	return nil
}
