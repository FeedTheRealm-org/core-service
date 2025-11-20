package models

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ModelsRepository defines the interface for models-related database operations.
type ModelsRepository interface {
	PublishModels(models []models.Model) ([]models.Model, error)
	GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error)
}
