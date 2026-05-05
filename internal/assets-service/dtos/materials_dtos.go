package dtos

import (
	"mime/multipart"

	"github.com/google/uuid"
)

type PutMaterialRequest struct {
	WorldID uuid.UUID
	UserID  uuid.UUID
	Files   []*multipart.FileHeader
}

type MaterialResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name,omitempty"`
	WorldID   uuid.UUID `json:"world_id,omitempty"`
	URL       string    `json:"url,omitempty"`
	CreatedAt string    `json:"created_at,omitempty"`
	UpdatedAt string    `json:"updated_at,omitempty"`
}
