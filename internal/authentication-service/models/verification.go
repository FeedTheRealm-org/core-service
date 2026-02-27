package models

import (
	"time"

	"github.com/google/uuid"
)

type AccountVerification struct {
	UserId           uuid.UUID `gorm:"primaryKey;constraint:OnDelete:CASCADE;"`
	VerificationCode string    `gorm:"not null;default:''"`
	Attempts         int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"default:now()"`
	ExpiresAt        time.Time `gorm:"default:now() + interval '10 minutes'"`

	User User `gorm:"foreignKey:UserId"`
}
