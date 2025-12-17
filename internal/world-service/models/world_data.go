package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type WorldData struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserId    uuid.UUID      `gorm:"not null"`
	Name      string         `gorm:"unique;not null"`
	Data      datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
}
