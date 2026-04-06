package models

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/dtos"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type ModelsService interface {
	UploadModels(modelsRequest dtos.BatchModelsRequest) ([]models.Model, error)
	GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error)
}
