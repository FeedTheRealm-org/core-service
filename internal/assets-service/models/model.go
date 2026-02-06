package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id         uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url        string                `gorm:"not null"`
	CategoryID uuid.UUID             `gorm:"type:uuid;not null;index"`
	WorldID    uuid.UUID             `gorm:"type:uuid;not null;index"`
	Category   ModelCategory         `gorm:"foreignKey:CategoryID;references:Id"`
	CreatedAt  time.Time             `gorm:"autoCreateTime"`
	UpdatedAt  time.Time             `gorm:"autoUpdateTime"`
	ModelFile  *multipart.FileHeader `gorm:"-" json:"-"`
}

func (Model) TableName() string {
	return "models"
}

type ModelCategory struct {
	Id        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name      string    `gorm:"not null;unique"`
	Models    []Model   `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ModelCategory) TableName() string {
	return "models_categories"
}
