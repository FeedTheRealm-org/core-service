package models

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	assetModels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ModelsRepository defines the interface for models-related database operations.
type ModelsRepository interface {
	UploadModel(model assetModels.Model) (*assetModels.Model, error)
	GetModelsByWorld(worldId uuid.UUID) ([]models.Model, error)
}
