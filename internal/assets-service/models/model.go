package models

import (
	"time"

	"github.com/google/uuid"
)

type Model struct {
	WorldID   uuid.UUID `gorm:"type:uuid;not null;primaryKey;uniqueIndex:idx_models_id_world_id,priority:2"`
	Id        uuid.UUID `gorm:"type:uuid;not null;default:gen_random_uuid();primaryKey;uniqueIndex:idx_models_id_world_id,priority:1"`
	Url       string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	CreatedBy uuid.UUID `gorm:"type:uuid;not null"`
}

func (Model) TableName() string {
	return "models"
}

func (m *Model) ToString() string {
	return "Model{id: " + m.Id.String() + ", url: " + m.Url + ", worldId: " + m.WorldID.String() + ", createdAt: " + m.CreatedAt.String() + ", updatedAt: " + m.UpdatedAt.String() + ", createdBy: " + m.CreatedBy.String() + "}"
}
