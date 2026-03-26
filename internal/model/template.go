package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	if s == nil {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringSlice: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

type Template struct {
	gorm.Model
	UserID    uint        `json:"user_id" gorm:"not null;uniqueIndex:idx_user_template_name"`
	Name      string      `json:"name" gorm:"not null;uniqueIndex:idx_user_template_name"`
	Channel   Channel     `json:"channel" gorm:"type:varchar(20);not null"`
	Subject   string      `json:"subject"`
	Body      string      `json:"body" gorm:"type:text;not null"`
	Variables StringSlice `json:"variables" gorm:"type:jsonb;default:'[]'"`
}
