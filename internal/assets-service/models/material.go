package models

import (
	"time"

	"github.com/google/uuid"
)

type Material struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	WorldID   uuid.UUID `gorm:"type:uuid;not null;index:idx_materials_world_id" json:"world_id"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

func (Material) TableName() string {
	return "materials"
}
