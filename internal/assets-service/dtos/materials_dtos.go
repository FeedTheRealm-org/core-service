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
	Name      string    `json:"name"`
	WorldID   uuid.UUID `json:"world_id"`
	URL       string    `json:"url"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}
