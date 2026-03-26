package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type TemplateData map[string]string

func (t TemplateData) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

func (t *TemplateData) Scan(value any) error {
	if value == nil {
		*t = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan TemplateData: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, t)
}

type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSlack Channel = "slack"
	ChannelInApp Channel = "inapp"
)

type Status string

const (
	StatusPending   Status = "pending"
	StatusSent      Status = "sent"
	StatusFailed    Status = "failed"
	StatusScheduled Status = "scheduled"
	StatusCancelled Status = "cancelled"
)

type Notification struct {
	gorm.Model
	Recipient  string     `json:"recipient" gorm:"not null"`
	Channel    Channel    `json:"channel" gorm:"type:varchar(20);not null"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`
	Status     Status     `json:"status" gorm:"type:varchar(20);default:'pending'"`
	TemplateID *uint      `json:"template_id"`
	Template   *Template  `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
	TemplateData TemplateData `json:"template_data,omitempty" gorm:"type:jsonb"`
	ScheduleAt   *time.Time   `json:"schedule_at"`
	SentAt       *time.Time   `json:"sent_at"`
}
