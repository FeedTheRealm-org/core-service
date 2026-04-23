package dtos

import (
	"time"
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
	Slots           int       `json:"slots"`
	UsedSlots       int       `json:"used_slots"`
	Status          string    `json:"status"`
	NextBillingDate time.Time `json:"next_billing_date"`
	AmountDue       float64   `json:"amount_due"`
}

type InternalSlotsCheckResponse struct {
	Allowed   bool `json:"allowed"`
	FreeSlots int  `json:"free_slots"`
}

type InternalUpdateUsedSlotsRequest struct {
	Slots   int  `json:"slots" binding:"required"`
	AreUsed bool `json:"are_used"`
}

type PricingInfoResponse struct {
	PricePerSlot    float64   `json:"price_per_slot"`
	NextBillingDate time.Time `json:"next_billing_date"`
}
