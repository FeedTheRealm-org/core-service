package models

import (
	"time"

	"github.com/google/uuid"
)

// ItemSprite represents a sprite asset for game items.
// This is managed by assets-service but references ItemCategory from items-service.
type ItemSprite struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url       string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
