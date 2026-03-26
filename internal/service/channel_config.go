package service

import (
	"encoding/json"
	"fmt"

	"github.com/devblin/tuskira/internal/model"
	"github.com/devblin/tuskira/internal/repository"
	"github.com/google/uuid"
)

// ChannelConfigService manages channel configurations (email, slack, inapp).
// Each channel has its own config schema validated in processConfig.
type ChannelConfigService struct {
	repo *repository.ChannelConfigRepository
}

func NewChannelConfigService(repo *repository.ChannelConfigRepository) *ChannelConfigService {
	return &ChannelConfigService{repo: repo}
}

func (s *ChannelConfigService) Upsert(cfg *model.ChannelConfig) error {
	processed, err := s.processConfig(cfg.Channel, cfg.Config, cfg.UserID)
	if err != nil {
		return fmt.Errorf("invalid config for channel %s: %w", cfg.Channel, err)
	}
	cfg.Config = processed
	return s.repo.Upsert(cfg)
}

func (s *ChannelConfigService) GetByChannel(channel model.Channel, userID uint) (*model.ChannelConfig, error) {
	return s.repo.FindByChannel(channel, userID)
}

func (s *ChannelConfigService) List(userID uint) ([]model.ChannelConfig, error) {
	return s.repo.FindAll(userID)
}

func (s *ChannelConfigService) Delete(channel model.Channel, userID uint) error {
	return s.repo.Delete(channel, userID)
}

// getExistingConnectionID loads the current inapp config from DB to preserve the connection ID across updates.
func (s *ChannelConfigService) getExistingConnectionID(channel model.Channel, userID uint) string {
	existing, err := s.repo.FindByChannel(channel, userID)
	if err != nil || existing == nil {
		return ""
	}
	var cfg model.InAppChannelConfig
	if err := json.Unmarshal(existing.Config, &cfg); err != nil {
		return ""
	}
	return cfg.ConnectionID
}

// processConfig validates and normalizes channel-specific configuration.
func (s *ChannelConfigService) processConfig(channel model.Channel, data model.ChannelConfigData, userID uint) (model.ChannelConfigData, error) {
	switch channel {
	case model.ChannelEmail:
		var cfg model.EmailChannelConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("invalid email config: %w", err)
		}
		if len(cfg.Providers) == 0 {
			return nil, fmt.Errorf("at least one email provider is required")
		}
		for i, p := range cfg.Providers {
			if p.Provider == "sendgrid" {
				if p.APIKey == "" {
					return nil, fmt.Errorf("provider %d: api_key is required for sendgrid", i+1)
				}
			} else {
				if p.Host == "" {
					return nil, fmt.Errorf("provider %d: host is required", i+1)
				}
			}
			if p.From == "" {
				return nil, fmt.Errorf("provider %d: from is required", i+1)
			}
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
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("invalid inapp config: %w", err)
			}
		}
		if cfg.ConnectionID == "" {
			cfg.ConnectionID = s.getExistingConnectionID(channel, userID)
		}
		if cfg.ConnectionID == "" {
			cfg.ConnectionID = uuid.New().String()
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
