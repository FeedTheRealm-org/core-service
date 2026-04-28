package models

import (
	"time"

	"github.com/google/uuid"
)

type CreatorBalance struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;unique"`
	Balance   float64   `gorm:"type:numeric(10,2);not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
