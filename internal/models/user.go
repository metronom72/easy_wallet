package models

import (
	"gorm.io/datatypes"
	"time"
)

type User struct {
	ID        string                       `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Username  string                       `gorm:"unique;" json:"username"`
	Email     string                       `gorm:"unique;" json:"email"`
	Providers datatypes.JSONType[[]string] `gorm:"type:jsonb" json:"providers"`
	CreatedAt time.Time                    `json:"created_at"`
	UpdatedAt time.Time                    `json:"updated_at"`
}
