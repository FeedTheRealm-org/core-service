package models

import (
	"time"

	"github.com/google/uuid"
)

type Purchase struct {
	PlayerID     uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"player_id"`
	CosmeticID   uuid.UUID `gorm:"type:uuid;not null;primaryKey" json:"cosmetic_id"`
	PurchaseDate time.Time `gorm:"type:timestamptz;not null;default:now()" json:"purchase_date"`
}

func (Purchase) TableName() string {
	return "purchases"
}
