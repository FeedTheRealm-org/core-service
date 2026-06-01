package dtos

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type PutMaterialRequest struct {
	WorldID uuid.UUID
	UserID  uuid.UUID
	Files   []*multipart.FileHeader
}

type MaterialResponse struct {
	ID           uuid.UUID `json:"id"`
	MaterialType int       `json:"material_type"`
	Name         string    `json:"name,omitempty"`
	WorldID      uuid.UUID `json:"world_id,omitempty"`
	URL          string    `json:"url,omitempty"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
