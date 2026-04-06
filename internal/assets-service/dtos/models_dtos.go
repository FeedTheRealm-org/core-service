package dtos

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type ModelRequest struct {
	Id        uuid.UUID             `json:"model_id" binding:"required"`
	Url       string                `json:"url" binding:"required"`
	ModelFile *multipart.FileHeader `gorm:"-" json:"-"`
}

type BatchModelsRequest struct {
	WorldID   uuid.UUID      `json:"world_id" binding:"required"`
	CreatedBy uuid.UUID      `json:"created_by" binding:"required"`
	Models    []ModelRequest `json:"models" binding:"required,dive"`
}

type ModelResponse struct {
	ModelID uuid.UUID `json:"model_id"`
	Url     string    `json:"url"`
}

type ModelsListResponse struct {
	WorldID uuid.UUID       `gorm:"type:uuid;not null" json:"world_id" binding:"required"`
	List    []ModelResponse `json:"models"`
}
