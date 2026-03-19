package models

import (
	"time"

	"github.com/google/uuid"
)

type GemBalance struct {
	UserId    uuid.UUID `json:"user_id" gorm:"type:uuid;primaryKey"`
	Gems      int       `json:"gems" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (GemBalance) TableName() string {
	return "gem_balances"
}
