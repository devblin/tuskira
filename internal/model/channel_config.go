package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type ChannelConfigData json.RawMessage

func (c ChannelConfigData) Value() (driver.Value, error) {
	if c == nil {
		return "{}", nil
	}
	return string(c), nil
}

func (c *ChannelConfigData) Scan(value any) error {
	if value == nil {
		*c = ChannelConfigData("{}")
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan ChannelConfigData: expected []byte, got %T", value)
	}
	*c = bytes
	return nil
}

func (c ChannelConfigData) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("{}"), nil
	}
	return []byte(c), nil
}

func (c *ChannelConfigData) UnmarshalJSON(data []byte) error {
	*c = data
	return nil
}

type ChannelConfig struct {
	gorm.Model
	Channel Channel           `json:"channel" gorm:"type:varchar(20);uniqueIndex;not null"`
	Enabled bool              `json:"enabled" gorm:"default:false"`
	Config  ChannelConfigData `json:"config" gorm:"type:jsonb;default:'{}'"`
}

// Typed config structs for each channel

type EmailChannelConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	From     string `json:"from"`
	TLS      bool   `json:"tls"`
}

type SlackChannelConfig struct {
	BotToken string         `json:"bot_token"`
	Channels []SlackChannel `json:"channels"`
}

type SlackChannel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type InAppChannelConfig struct {
	ConnectionID string `json:"connection_id"`
}
