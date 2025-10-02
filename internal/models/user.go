package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username     string `gorm:"unique;size:100" json:"username"`
	Email        string `gorm:"uniqueIndex;size:100;not null" json:"email"`
	Password     string `gorm:"not null" json:"-"`
	TokenVersion int    `gorm:"not null;default:0"`
}
