package models

import (
	"time"

	"github.com/google/uuid"
)

// ItemCategory represents a category for items and their sprites.
// This is separate from the Category table in assets-service.
type ItemCategory struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
