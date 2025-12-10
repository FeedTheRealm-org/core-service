package models

import (
	"time"

	"github.com/google/uuid"
)

// ItemCategory is a reference model for the ItemCategory table in items-service.
// This table is NOT managed by assets-service, but we need the model for validation.
// NOTE: This should match the ItemCategory model in items-service exactly.
type ItemCategory struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

// TableName specifies the table name for ItemCategory
func (ItemCategory) TableName() string {
	return "item_categories"
}
