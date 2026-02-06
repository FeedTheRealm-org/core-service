package models

import (
	"time"

	"github.com/google/uuid"
)

// Item represents a sprite asset for game items.
// This is managed by assets-service but references ItemCategory from items-service.
type Item struct {
	Id         uuid.UUID    `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url        string       `gorm:"not null"`
	CategoryID uuid.UUID    `gorm:"type:uuid;not null;index"`
	Category   ItemCategory `gorm:"foreignKey:CategoryID;references:Id"`
	CreatedAt  time.Time    `gorm:"autoCreateTime"`
	UpdatedAt  time.Time    `gorm:"autoUpdateTime"`
}

func (Item) TableName() string {
	return "items"
}

type ItemCategory struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	Items     []Item    `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ItemCategory) TableName() string {
	return "items_categories"
}
