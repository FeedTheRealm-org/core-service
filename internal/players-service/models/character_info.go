package models

import (
	"time"

	"github.com/google/uuid"
)

type CharacterInfo struct {
	UserId        uuid.UUID `gorm:"type:uuid;primaryKey"`
	CharacterName string    `gorm:"unique;not null"`
	CharacterBio  string    `gorm:"not null"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
