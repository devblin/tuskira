package model

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `json:"email" gorm:"uniqueIndex;not null"`
	Password string `json:"-" gorm:"not null"`
}
