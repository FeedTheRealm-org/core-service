package models

import (
	"time"
)

type GemMetrics struct {
	ID          int32     `json:"id" gorm:"primaryKey"`
	GemsBought  int64     `json:"gems_bought" gorm:"not null;default:0"`
	GemsSpent   int64     `json:"gems_spent" gorm:"not null;default:0"`
	GemsRevenue float64   `json:"gems_revenue" gorm:"type:numeric(12,2);not null;default:0"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (GemMetrics) TableName() string {
	return "gem_metrics"
}
