package world

import (
	"errors"

	"github.com/FeedTheRealm-org/core-service/config"
	world_errors "github.com/FeedTheRealm-org/core-service/internal/world-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type worldRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewWorldRepository creates a new instance of WorldRepository.
func NewWorldRepository(conf *config.Config, db *config.DB) WorldRepository {
	return &worldRepository{
		conf: conf,
		db:   db,
	}
}

// StoreWorldData stores new world data in the database.
func (r *worldRepository) StoreWorldData(newWorldData *models.WorldData) error {
	return r.db.Conn.Save(newWorldData).Error
}

// GetWorldData retrieves information for a specific world by ID.
func (r *worldRepository) GetWorldData(worldID uuid.UUID) (*models.WorldData, error) {
	var wd models.WorldData
	err := r.db.Conn.Where("id = ?", worldID).First(&wd).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, world_errors.NewWorldNotFound(err.Error())
		}
		return nil, err
	}
	return &wd, nil
}

// GetWorldsList retrieves a paginated list of worlds.
func (r *worldRepository) GetWorldsList(offset int, limit int) ([]*models.WorldData, error) {
	var worlds []*models.WorldData
	err := r.db.Conn.Offset(offset).Limit(limit).Find(&worlds).Error
	return worlds, err
}
