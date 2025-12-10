package dtos

import (
	"github.com/google/uuid"
)

// BundlePublishResponse represents the response when a bundle is successfully uploaded
type BundlePublishResponse struct {
	WorldID   uuid.UUID `json:"world_id" example:"1a2b3c4d-5e6f-7890-abcd-ef1234567890"`
	BundleURL string    `json:"bundle_url" example:"bucket/bundles/1a2b3c4d-5e6f-7890-abcd-ef1234567890/bundle.zip"`
}

// BundleDownloadResponse represents bundle file metadata
type BundleDownloadResponse struct {
	WorldID   uuid.UUID `json:"world_id"`
	BundleURL string    `json:"bundle_url"`
}
