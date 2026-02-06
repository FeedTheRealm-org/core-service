package models

import (
	"time"

	"github.com/google/uuid"
)

type Cosmetic struct {
	Id         uuid.UUID        `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url        string           `gorm:"not null"`
	CategoryID uuid.UUID        `gorm:"type:uuid;not null;index"`
	Category   CosmeticCategory `gorm:"foreignKey:CategoryID;references:Id"`
	CreatedAt  time.Time        `gorm:"autoCreateTime"`
	UpdatedAt  time.Time        `gorm:"autoUpdateTime"`
}

func (Cosmetic) TableName() string {
	return "cosmetics"
}

type CosmeticCategory struct {
	Id        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string     `gorm:"not null;unique"`
	Cosmetics []Cosmetic `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime"`
}

func (CosmeticCategory) TableName() string {
	return "cosmetics_categories"
}
