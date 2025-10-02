package models

import "time"

type Rate struct {
	Currency  string    `gorm:"primaryKey;size:3"`
	Rate      float64   `gorm:"not null"`
	UpdatedAt time.Time `gorm:"index"`
}
