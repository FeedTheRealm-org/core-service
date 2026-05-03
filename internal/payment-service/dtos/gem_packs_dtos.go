package dtos

import (
	"time"

	"github.com/google/uuid"
)

type CreateGemPackRequest struct {
	Name  string  `json:"name" binding:"required"`
	Gems  int     `json:"gems" binding:"required"`
	Price float64 `json:"price" binding:"required"`
}

type UpdateGemPackRequest struct {
	Name  string  `json:"name,omitempty"`
	Gems  int     `json:"gems,omitempty"`
	Price float64 `json:"price,omitempty"`
}

type GemPackResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Gems      int       `json:"gems"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GemPackDeletedResponse struct {
	Id uuid.UUID `json:"id"`
}
