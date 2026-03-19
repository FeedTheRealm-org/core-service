package dtos

import "github.com/google/uuid"

type GemBalanceResponse struct {
	UserId uuid.UUID `json:"user_id"`
	Gems   int       `json:"gems"`
}

type UpdateGemBalanceRequest struct {
	UserId uuid.UUID `json:"user_id"`
	Gems   int       `json:"gems" binding:"required"`
}

type CheckoutRequest struct {
	GemPackId  uuid.UUID `json:"gem_pack_id" binding:"required"`
	SuccessUrl string    `json:"success_url" binding:"required"`
	CancelUrl  string    `json:"cancel_url" binding:"required"`
}

type CheckoutResponse struct {
	CheckoutUrl string `json:"checkout_url"`
}

type WebhookResponse struct {
}
