package models

import (
	"time"

	"github.com/google/uuid"
)

type Category struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	Sprites   []Sprite  `gorm:"many2many:sprite_categories;"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
