package dtos

import "github.com/google/uuid"

type AddCosmeticCategoryRequest struct {
	CategoryName string `json:"category_name" binding:"required"`
}

type CosmeticCategoryResponse struct {
	CategoryId   uuid.UUID `json:"category_id"`
	CategoryName string    `json:"category_name"`
}

type CosmeticResponse struct {
	CosmeticId  uuid.UUID `json:"cosmetic_id"`
	CosmeticUrl string    `json:"cosmetic_url"`
}

// CosmeticCategoryListResponse returns a list of sprite categories.
type CosmeticCategoryListResponse struct {
	CategoryList []CosmeticCategoryResponse `json:"category_list"`
}

// CosmeticsListResponse returns a list of sprites.
type CosmeticsListResponse struct {
	CosmeticsList []CosmeticResponse `json:"cosmetics_list"`
}
