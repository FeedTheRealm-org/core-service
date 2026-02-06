package dtos

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// UploadItemSpritesRequest representa el form-data para carga múltiple de sprites de items.
type UploadItemSpritesRequest struct {
	Ids     []uuid.UUID             `form:"ids[]" binding:"required"`
	Sprites []*multipart.FileHeader `form:"sprites[]" binding:"required"`
}

type AddItemCategoryRequest struct {
	CategoryName string `json:"category_name" binding:"required"`
}

type ItemCategoryResponse struct {
	CategoryId   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
}

// ItemResponse represents an item sprite response.
type ItemResponse struct {
	Id        uuid.UUID `json:"id"`
	Url       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ItemListResponse represents a list of item sprites.
type ItemListResponse struct {
	Items []ItemResponse `json:"items"`
}
