package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	WorldID     uuid.UUID       `gorm:"type:uuid;not null;index"`
	ModelID     uuid.UUID       `gorm:"type:uuid;not null"`
	Categories  []ModelCategory `gorm:"many2many:model_categories;"`
	Name        string          `gorm:"type:text;not null"`
	ModelURL    string          `gorm:"type:text;not null"`
	MaterialURL string          `gorm:"type:text"`
	CreatedAt   time.Time       `gorm:"autoCreateTime"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime"`
	// These fields are used only during upload and are not stored in DB
	ModelFile    *multipart.FileHeader `gorm:"-" json:"-"`
	MaterialFile *multipart.FileHeader `gorm:"-" json:"-"`
}
