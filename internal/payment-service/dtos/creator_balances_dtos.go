package dtos

import "github.com/google/uuid"

type CreatorBalanceResponse struct {
	UserID  uuid.UUID `json:"user_id"`
	Balance float64   `json:"balance"`
}
