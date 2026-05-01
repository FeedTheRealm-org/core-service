package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v85"
)

type ZonesSubscriptions struct {
	ID                   uuid.UUID                 `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID               uuid.UUID                 `json:"user_id" gorm:"type:uuid;uniqueIndex"`
	StripeCustomerID     string                    `json:"stripe_customer_id" gorm:"type:varchar(255);not null"`
	StripeSubscriptionID string                    `json:"stripe_subscription_id" gorm:"type:varchar(255)"`
	TotalSlots           int                       `json:"total_slots" gorm:"type:int;not null;default:0"`
	UsedSlots            int                       `json:"used_slots" gorm:"type:int;not null;default:0"`
	AmountDue            decimal.Decimal           `json:"amount_due" gorm:"type:decimal(10,2);not null"`
	Status               stripe.SubscriptionStatus `json:"status" gorm:"type:varchar(50);not null;default:'inactive'"`
	NextBillingDate      time.Time                 `json:"next_billing_date"`
	CreatedAt            time.Time                 `json:"created_at"`
	UpdatedAt            time.Time                 `json:"updated_at"`
}

func (ZonesSubscriptions) TableName() string {
	return "zones_subscriptions"
}
