package models

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assetModels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type modelsRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewModelsRepository creates a new instance of ModelsRepository.
func NewModelsRepository(conf *config.Config, db *config.DB) ModelsRepository {
	return &modelsRepository{
		conf: conf,
		db:   db,
	}
}

func (mr *modelsRepository) UploadModels(modelsList []assetModels.Model) ([]assetModels.Model, error) {

	logger.Logger.Infof("REPO: Uploading %d models to the database", len(modelsList))

	for _, model := range modelsList {
		result := mr.db.Conn.Save(&model)
		if result.Error != nil {
			logger.Logger.Errorf("REPO: Failed to upload model %s: %v", model.Id, result.Error)
			return nil, result.Error
		}
		logger.Logger.Infof("REPO: Model uploaded (rows affected: %d): %s", result.RowsAffected, model.ToString())
	}

	logger.Logger.Infof("REPO: Published %d models to the db", len(modelsList))
	return modelsList, nil
}

func (mr *modelsRepository) GetModelsByWorld(worldId uuid.UUID) ([]assetModels.Model, error) {
	var modelsList []assetModels.Model
	if err := mr.db.Conn.Where("world_id = ?", worldId).Find(&modelsList).Error; err != nil {
		return nil, err
	}
	return modelsList, nil
}
