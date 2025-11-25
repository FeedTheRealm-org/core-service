package models

import (
	"time"

	"github.com/google/uuid"
)

// ItemSprite represents a sprite asset for game items.
type ItemSprite struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Category  string    `gorm:"not null"` // e.g., "armor", "weapon", "consumable"
	Url       string    `gorm:"not null"` // File path or URL to the sprite
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
