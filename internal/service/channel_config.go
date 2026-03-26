package service

import (
	"encoding/json"
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/google/uuid"
)

type ChannelConfigService struct {
	repo *repository.ChannelConfigRepository
}

func NewChannelConfigService(repo *repository.ChannelConfigRepository) *ChannelConfigService {
	return &ChannelConfigService{repo: repo}
}

func (s *ChannelConfigService) Upsert(cfg *model.ChannelConfig) error {
	processed, err := s.processConfig(cfg.Channel, cfg.Config)
	if err != nil {
		return fmt.Errorf("invalid config for channel %s: %w", cfg.Channel, err)
	}
	cfg.Config = processed
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

func (s *ChannelConfigService) processConfig(channel model.Channel, data model.ChannelConfigData) (model.ChannelConfigData, error) {
	switch channel {
	case model.ChannelEmail:
		var cfg model.EmailChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid email config: %w", err)
		}
		if cfg.Host == "" {
			return nil, fmt.Errorf("host is required")
		}
		if cfg.From == "" {
			return nil, fmt.Errorf("from is required")
		}
		return data, nil
	case model.ChannelSlack:
		var cfg model.SlackChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid slack config: %w", err)
		}
		if cfg.BotToken == "" {
			return nil, fmt.Errorf("bot_token is required")
		}
		return data, nil
	case model.ChannelInApp:
		var cfg model.InAppChannelConfig
		if data != nil {
			_ = json.Unmarshal(data, &cfg)
		}
		if cfg.ConnectionID == "" {
			existing, err := s.repo.FindByChannel(channel)
			if err == nil && existing != nil {
				var existingCfg model.InAppChannelConfig
				if err := json.Unmarshal(existing.Config, &existingCfg); err == nil && existingCfg.ConnectionID != "" {
					cfg.ConnectionID = existingCfg.ConnectionID
				}
			}
			if cfg.ConnectionID == "" {
				cfg.ConnectionID = uuid.New().String()
			}
		}
		processed, err := json.Marshal(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal inapp config: %w", err)
		}
		return model.ChannelConfigData(processed), nil
	default:
		return nil, fmt.Errorf("unknown channel: %s", channel)
	}
}
