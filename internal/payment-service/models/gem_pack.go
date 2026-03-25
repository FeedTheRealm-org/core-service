package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type GemPack struct {
	Id        uuid.UUID       `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string          `json:"name"`
	Gems      int             `json:"gems"`
	Price     decimal.Decimal `json:"price" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

func (GemPack) TableName() string {
	return "gem_packs"
}
