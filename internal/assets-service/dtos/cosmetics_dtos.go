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

type InternalCosmeticResponse struct {
	CosmeticId    uuid.UUID `json:"cosmetic_id"`
	CosmeticPrice float64   `json:"cosmetic_price"`
}

// CosmeticCategoryListResponse returns a list of sprite categories.
type CosmeticCategoryListResponse struct {
	CategoryList []CosmeticCategoryResponse `json:"category_list"`
}

// CosmeticsListResponse returns a list of sprites.
type CosmeticsListResponse struct {
	CosmeticsList []CosmeticResponse `json:"cosmetics_list"`
	TotalCount    int64              `json:"total_count"`
}

type InternalPurchaseCosmeticForUserRequest struct {
	CosmeticId uuid.UUID `json:"cosmetic_id" binding:"required"`
}

type InternalPurchaseCosmeticForUserResponse struct {
	UserId     uuid.UUID `json:"user_id"`
	CosmeticId uuid.UUID `json:"cosmetic_id"`
}
