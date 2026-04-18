package dtos

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v84"
)

type SubscriptionRequest struct {
	Slots int    `json:"slots" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type CheckoutSessionRequest struct {
	Slots      int    `json:"slots" validate:"required"`
	SuccessUrl string `json:"success_url" validate:"required"`
	CancelUrl  string `json:"cancel_url" validate:"required"`
}

type UpdateSubscriptionRequest struct {
	Slots int `json:"slots" validate:"required"`
}

type SubscriptionStatusResponse struct {
	Slots           int                       `json:"slots"`
	UsedSlots       int                       `json:"used_slots"`
	Status          stripe.SubscriptionStatus `json:"status"` // active, trailing, canceled
	NextBillingDate time.Time                 `json:"next_billing_date"`
	AmountDue       decimal.Decimal           `json:"amount_due"`
}

type InternalSlotsCheckRequest struct {
	RequiredSlots int `json:"required_slots" validate:"required"`
}

type InternalSlotsCheckResponse struct {
	Allowed    bool `json:"allowed"`
	TotalSlots int  `json:"total_slots"`
}
