package models

import (
	"time"

	"github.com/google/uuid"
)

type CreatorBalance struct {
	UserID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	Balance   float64   `gorm:"type:numeric(10,2);not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
