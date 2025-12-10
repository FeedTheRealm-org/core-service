package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

// TODO: currently, all world models, including their materials, are stored in a bundle.
// In the future, we might want to separate it into multiple chunk files, as in, one bundle per world chunk.
type Bundle struct {
	WorldID   uuid.UUID `gorm:"column:world_id;type:uuid;not null;primaryKey" json:"world_id"`
	BundleURL string    `gorm:"column:bundle_url;type:text;not null" json:"bundle_url"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	// This field is used only during upload and is not stored in DB
	BundleFile *multipart.FileHeader `gorm:"-" json:"-"`
}
