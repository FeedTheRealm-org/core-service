package models

import (
	"time"

	"github.com/google/uuid"
)

// WorldJoinToken stores a one-time token used to safely resolve a user ID on game-server join.
type WorldJoinToken struct {
	TokenId    uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserId     uuid.UUID  `gorm:"type:uuid;not null;index"`
	WorldId    string     `gorm:"not null;index"`
	ExpiresAt  time.Time  `gorm:"not null;index"`
	ConsumedAt *time.Time `gorm:"index"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime"`
}
