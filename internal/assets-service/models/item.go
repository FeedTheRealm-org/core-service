package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a sprite asset for game items.
type Item struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url       string    `gorm:"not null"`
	WorldID   uuid.UUID `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
}

func (Item) TableName() string {
	return "items"
}
