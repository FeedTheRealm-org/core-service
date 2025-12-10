package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a game item with its metadata.
type Item struct {
	Id          uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string       `gorm:"not null"`
	Description string       `gorm:"not null"`
	CategoryId  uuid.UUID    `gorm:"type:uuid;not null"`
	Category    ItemCategory `gorm:"foreignKey:CategoryId;constraint:OnDelete:RESTRICT;"`
	SpriteId    uuid.UUID    `gorm:"type:uuid;not null"`
	CreatedAt   time.Time    `gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"autoUpdateTime"`
}
