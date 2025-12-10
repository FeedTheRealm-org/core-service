package bundles

import (
	"github.com/FeedTheRealm-org/core-service/config"
	assetModels "github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type bundlesRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewBundlesRepository creates a new instance of BundlesRepository.
func NewBundlesRepository(conf *config.Config, db *config.DB) BundlesRepository {
	return &bundlesRepository{
		conf: conf,
		db:   db,
	}
}

func (mr *bundlesRepository) PublishWorldBundle(bundle assetModels.Bundle) (assetModels.Bundle, error) {
	if err := mr.db.Conn.Create(&bundle).Error; err != nil {
		return assetModels.Bundle{}, err
	}
	return bundle, nil
}

func (mr *bundlesRepository) GetWorldBundle(worldId uuid.UUID) (assetModels.Bundle, error) {
	var bundle assetModels.Bundle
	if err := mr.db.Conn.Where("world_id = ?", worldId).Take(&bundle).Error; err != nil {
		return assetModels.Bundle{}, err
	}
	return bundle, nil
}
