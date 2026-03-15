package dtos

import (
	"time"

	"github.com/google/uuid"
)

type CreatePackRequest struct {
	Name  string  `json:"name" binding:"required"`
	Gems  int     `json:"gems" binding:"required"`
	Price float32 `json:"price" binding:"required"`
}

type UpdatePackRequest struct {
	Name  string  `json:"name,omitempty"`
	Gems  int     `json:"gems,omitempty"`
	Price float32 `json:"price,omitempty"`
}

type PackResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Gems      int       `json:"gems"`
	Price     float32   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PackDeletedResponse struct {
	Id uuid.UUID `json:"id"`
}
