package models

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WorldZone struct {
	ID       int            `gorm:"not null;primaryKey"`
	WorldID  uuid.UUID      `gorm:"type:uuid;not null;primaryKey"`
	ZoneData datatypes.JSON `gorm:"type:jsonb;not null"`
	IsActive bool           `gorm:"not null;default:false"`
}
