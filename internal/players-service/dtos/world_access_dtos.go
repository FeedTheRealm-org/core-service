package dtos

import (
	"time"

	"github.com/google/uuid"
)

type IssueWorldJoinTokenRequest struct {
	WorldId string `json:"world_id" binding:"required"`
}

type IssueWorldJoinTokenResponse struct {
	TokenId   uuid.UUID `json:"token_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ConsumeWorldJoinTokenRequest struct {
	TokenId string `json:"token_id" binding:"required"`
}

type ConsumeWorldJoinTokenResponse struct {
	UserId  uuid.UUID `json:"user_id"`
	WorldId string    `json:"world_id"`
}
