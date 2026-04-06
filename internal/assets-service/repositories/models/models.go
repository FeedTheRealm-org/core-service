package models

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assetModels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"gorm.io/gorm/clause"
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
	tx := mr.db.Conn.Begin()

	logger.Logger.Infof("REPO: Uploading %d models to the database", len(modelsList))

	for _, model := range modelsList {

		if err := tx.
			Clauses(
				clause.OnConflict{
					Columns:   []clause.Column{{Name: "id"}},
					DoUpdates: clause.AssignmentColumns([]string{"url", "updated_at"}),
				},
			).Create(&model).Error; err != nil {
			tx.Rollback()
			return nil, err
		}

		logger.Logger.Infof("REPO: Model uploaded: %s", model.ToString())
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
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
