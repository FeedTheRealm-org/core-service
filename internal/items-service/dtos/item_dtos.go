package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateItemRequest represents the request payload for creating a single item.
type CreateItemRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	Category    string    `json:"category" binding:"required"`
	SpriteId    uuid.UUID `json:"sprite_id" binding:"required"`
}

// CreateItemBatchRequest represents the request payload for creating multiple items.
type CreateItemBatchRequest struct {
	Items []CreateItemRequest `json:"items" binding:"required"`
}

// ItemMetadataResponse represents a single item metadata response.
type ItemMetadataResponse struct {
	Id          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	SpriteId    uuid.UUID `json:"sprite_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ItemsListResponse represents the response payload for retrieving all items metadata.
type ItemsListResponse struct {
	Items []ItemMetadataResponse `json:"items"`
}

// UploadSpriteRequest represents the request payload for uploading a sprite.
// This is a multipart form request, so we use form tags.
type UploadSpriteRequest struct {
	Category string `form:"category" binding:"required"`
}

// ItemSpriteResponse represents a sprite metadata response.
type ItemSpriteResponse struct {
	Id        uuid.UUID `json:"id"`
	Category  string    `json:"category"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
