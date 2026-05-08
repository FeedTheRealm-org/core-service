package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type MaterialType int

const (
	GroundMaterial MaterialType = iota
	SkyBoxMaterial
)

type Material struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	WorldID   uuid.UUID `gorm:"type:uuid;primaryKey;index:idx_materials_world_id" json:"world_id"`
	Name      string    `gorm:"type:text;not null" json:"name"`
	URL       string    `gorm:"type:text;not null" json:"url"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null" json:"created_by"`
}

func (Material) TableName() string {
	return "materials"
}

func ParseMaterialType(i int) (MaterialType, error) {
	switch i {
	case 0:
		return GroundMaterial, nil
	case 1:
		return SkyBoxMaterial, nil
	default:
		return 0, errors.New("invalid material type: " + string(rune(i)))
	}
}
