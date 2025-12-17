package models

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assetModels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
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

func (mr *modelsRepository) PublishModels(modelsList []assetModels.Model) ([]assetModels.Model, error) {
	tx := mr.db.Conn.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var publishedModels = make([]assetModels.Model, 0, len(modelsList))
	for _, model := range modelsList {
		if err := tx.Create(&model).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		publishedModels = append(publishedModels, model)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return publishedModels, nil
}

func (mr *modelsRepository) GetModelsByWorld(worldId uuid.UUID) ([]assetModels.Model, error) {
	var modelsList []assetModels.Model
	if err := mr.db.Conn.Where("world_id = ?", worldId).Find(&modelsList).Error; err != nil {
		return nil, err
	}
	return modelsList, nil
}
