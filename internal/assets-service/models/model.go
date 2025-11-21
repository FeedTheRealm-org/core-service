package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModelID     uuid.UUID `gorm:"type:uuid;not null"`
	WorldID     uuid.UUID `gorm:"type:uuid;not null"`
	Name        string    `gorm:"type:text;not null"`
	ModelURL    string    `gorm:"type:text;not null"`
	MaterialURL string    `gorm:"type:text;not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	// These fields are used only during upload and are not stored in DB
	ModelFile    *multipart.FileHeader `gorm:"-" json:"-"`
	MaterialFile *multipart.FileHeader `gorm:"-" json:"-"`
}
