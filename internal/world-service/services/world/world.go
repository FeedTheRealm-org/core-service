package world

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	"github.com/google/uuid"
)

type worldService struct {
	conf            *config.Config
	worldRepository world.WorldRepository
}

// NewWorldService creates a new instance of WorldService.
func NewWorldService(conf *config.Config, worldRepository world.WorldRepository) WorldService {
	return &worldService{
		conf:            conf,
		worldRepository: worldRepository,
	}
}

func (cs *worldService) PublishWorld(newWorldData *models.WorldData) (*models.WorldData, error) {
	return cs.worldRepository.StoreWorldData(newWorldData)
}

func (cs *worldService) GetWorld(worldID uuid.UUID) (*models.WorldData, error) {
	return cs.worldRepository.GetWorldData(worldID)
}

func (cs *worldService) GetWorldsList(offset int, limit int) ([]*models.WorldData, error) {
	return cs.worldRepository.GetWorldsList(offset, limit)
}

func (cs *worldService) ClearDatabase() error {
	return cs.worldRepository.ClearDatabase()
}
