package models

import (
	"mime/multipart"
	"time"

	"github.com/google/uuid"
)

type Model struct {
	Id        uuid.UUID             `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Url       string                `gorm:"not null"`
	WorldID   uuid.UUID             `gorm:"type:uuid;not null;index"`
	CreatedAt time.Time             `gorm:"autoCreateTime"`
	UpdatedAt time.Time             `gorm:"autoUpdateTime"`
	ModelFile *multipart.FileHeader `gorm:"-" json:"-"`
	CreatedBy uuid.UUID             `gorm:"type:uuid;not null"`
}

func (Model) TableName() string {
	return "models"
}

func (m *Model) ToString() string {
	return "Model{id: " + m.Id.String() + ", url: " + m.Url + ", worldId: " + m.WorldID.String() + ", createdAt: " + m.CreatedAt.String() + ", updatedAt: " + m.UpdatedAt.String() + ", createdBy: " + m.CreatedBy.String() + "}"
}
