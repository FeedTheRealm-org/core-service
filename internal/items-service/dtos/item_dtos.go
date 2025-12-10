package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateItemRequest represents the request payload for creating a single item.
type CreateItemRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description" binding:"required"`
	CategoryId  uuid.UUID `json:"category_id" binding:"required"`
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
	CategoryId  uuid.UUID `json:"category_id"`
	SpriteId    uuid.UUID `json:"sprite_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ItemsListResponse represents the response payload for retrieving all items metadata.
type ItemsListResponse struct {
	Items []ItemMetadataResponse `json:"items"`
}

// CreateItemCategoryRequest represents the request payload for creating an item category.
type CreateItemCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// ItemCategoryResponse represents an item category response.
type ItemCategoryResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemCategoriesListResponse represents the response payload for retrieving all item categories.
type ItemCategoriesListResponse struct {
	Categories []ItemCategoryResponse `json:"categories"`
}
