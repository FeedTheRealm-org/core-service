package dtos

import (
	"time"

	"github.com/google/uuid"
)

// UploadItemSpriteRequest represents the multipart form request for uploading item sprites.
type UploadItemSpriteRequest struct {
	CategoryId uuid.UUID `form:"category_id" binding:"required"`
}

// ItemSpriteResponse represents an item sprite response.
type ItemSpriteResponse struct {
	Id        uuid.UUID `json:"id"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemSpritesListResponse represents a list of item sprites.
type ItemSpritesListResponse struct {
	Sprites []ItemSpriteResponse `json:"sprites"`
}
