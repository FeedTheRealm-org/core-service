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
