package dtos

import (
	"github.com/google/uuid"
)

type ModelResponse struct {
	ModelID uuid.UUID `json:"model_id"`
	Name    string    `json:"name"`
}

type ModelsListResponse struct {
	WorldID uuid.UUID       `gorm:"type:uuid;not null" json:"world_id" binding:"required"`
	List    []ModelResponse `json:"models"`
}
