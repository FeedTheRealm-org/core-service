package dtos

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type ModelRequest struct {
	WorldID   uuid.UUID             `json:"world_id" binding:"required"`
	Id        uuid.UUID             `json:"model_id" binding:"required"`
	CreatedBy uuid.UUID             `json:"created_by" binding:"required"`
	Url       string                `json:"url" binding:"required"`
	ModelFile *multipart.FileHeader `gorm:"-" json:"-"`
}

type ModelResponse struct {
	ModelID   uuid.UUID `json:"model_id"`
	Url       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ModelsListResponse struct {
	WorldID uuid.UUID       `gorm:"type:uuid;not null" json:"world_id" binding:"required"`
	List    []ModelResponse `json:"models"`
}
