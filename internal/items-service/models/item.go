package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a game item with its metadata.
type Item struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `gorm:"not null"`
	Description string    `gorm:"not null"`
	SpriteId    uuid.UUID `gorm:"type:uuid;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
}
