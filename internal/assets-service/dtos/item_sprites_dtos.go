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
	Id         uuid.UUID `json:"id"`
	CategoryId uuid.UUID `json:"category_id"`
	Url        string    `json:"url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ItemSpritesListResponse represents a list of item sprites.
type ItemSpritesListResponse struct {
	Sprites []ItemSpriteResponse `json:"sprites"`
}

// ItemCategoryResponse represents an item category (read from items-service).
type ItemCategoryResponse struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemCategoriesListResponse represents a list of item categories.
type ItemCategoriesListResponse struct {
	Categories []ItemCategoryResponse `json:"categories"`
}
