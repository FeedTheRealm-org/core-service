package dtos

import "github.com/google/uuid"

type AddSpriteCategoryRequest struct {
	CategoryName string `json:"category_name" binding:"required"`
}

type SpriteCategoryResponse struct {
	CategoryID   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
}

type SpriteResponse struct {
	SpriteID  uuid.UUID `json:"sprite_id"`
	SpriteUrl string    `json:"sprite_url"`
}

// SpriteCategoryListResponse returns a list of sprite categories.
type SpriteCategoryListResponse struct {
	CategoryList []SpriteCategoryResponse `json:"category_list"`
}

// SpritesListResponse returns a list of sprites.
type SpritesListResponse struct {
	SpritesList []SpriteResponse `json:"sprites_list"`
}
