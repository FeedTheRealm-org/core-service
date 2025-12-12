package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateItemRequest represents the request payload for creating a single item.
type CreateItemRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	SpriteId    uuid.UUID `json:"sprite_id"`
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
	SpriteId    uuid.UUID `json:"sprite_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UpdateItemSpriteRequest represents the request payload for updating the sprite associated to an item.
type UpdateItemSpriteRequest struct {
	SpriteId uuid.UUID `json:"sprite_id" binding:"required"`
}

// ItemsListResponse represents the response payload for retrieving all items metadata.
type ItemsListResponse struct {
	Items []ItemMetadataResponse `json:"items"`
}

// (Item categories removed from API)
