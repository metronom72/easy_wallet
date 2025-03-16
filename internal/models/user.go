package models

import (
	"time"
)

type User struct {
	ID         string `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username   string `gorm:"unique;" json:"username"`
	Email      string `gorm:"unique;" json:"email"`
	ExternalID string `gorm:"not null;" json:"external_id"`
	Provider   string `gorm:"not null;" json:"provider"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
