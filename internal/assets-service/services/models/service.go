package models

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type ModelsService interface {
	PublishModels(worldId uuid.UUID, models []models.Model) ([]models.Model, error)
	GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error)
}
