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

type ProcessedStripeWebhookEvent struct {
	EventID   string    `json:"event_id" gorm:"column:event_id;primaryKey"`
	SessionID string    `json:"session_id" gorm:"column:session_id;uniqueIndex;not null"`
	EventType string    `json:"event_type" gorm:"column:event_type;not null"`
	UserID    uuid.UUID `json:"user_id" gorm:"column:user_id;type:uuid;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

func (ProcessedStripeWebhookEvent) TableName() string {
	return "processed_stripe_webhook_events"
}
