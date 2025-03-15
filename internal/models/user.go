package models

import (
	"gorm.io/datatypes"
	"time"
)

type User struct {
	ID        uint                         `gorm:"primaryKey" json:"id"`
	Username  string                       `gorm:"unique;not null" json:"username"`
	Email     string                       `gorm:"unique;not null" json:"email"`
	Password  string                       `gorm:"not null" json:"-"`
	Providers datatypes.JSONType[[]string] `gorm:"type:jsonb" json:"providers"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}
