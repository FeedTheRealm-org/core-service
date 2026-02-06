package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a sprite asset for game items.
// This is managed by assets-service but references ItemCategory from items-service.
type Item struct {
	Id         uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url        string         `gorm:"not null"`
	Categories []ItemCategory `gorm:"many2many:item_categories;"`
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
}
