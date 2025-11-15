package models

import (
	"time"

	"github.com/google/uuid"
)

type Sprite struct {
	Id         uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url        string     `gorm:"not null"`
	Categories []Category `gorm:"many2many:sprite_categories;"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}
