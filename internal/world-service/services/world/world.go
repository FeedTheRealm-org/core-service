package world

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/google/uuid"
)

type worldService struct {
	conf                  *config.Config
	worldRepository       world.WorldRepository
	serverRegistryService server_registry.ServerRegistryService
}

// NewWorldService creates a new instance of WorldService.
func NewWorldService(
	conf *config.Config,
	worldRepository world.WorldRepository,
	serverRegistryService server_registry.ServerRegistryService,
) WorldService {
	return &worldService{
		conf:                  conf,
		worldRepository:       worldRepository,
		serverRegistryService: serverRegistryService,
	}
}

func (cs *worldService) PublishWorld(newWorldData *models.WorldData) (*models.WorldData, error) {
	createdWorld, err := cs.worldRepository.StoreWorldData(newWorldData)
	if err != nil {
		return nil, err
	}

	const defaultZoneID = 1
	if err := cs.serverRegistryService.StartNewJob(createdWorld.ID, defaultZoneID); err != nil {
		return nil, err
	}

	return createdWorld, nil
}

func (cs *worldService) GetWorld(worldID uuid.UUID) (*models.WorldData, error) {
	return cs.worldRepository.GetWorldData(worldID)
}

func (cs *worldService) UpdateWorld(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error) {
	updatedWorld, err := cs.worldRepository.UpdateWorldData(worldID, userId, data, description)
	if err != nil {
		return nil, err
	}

	const defaultZoneID = 1
	if err := cs.serverRegistryService.StartNewJob(worldID, defaultZoneID); err != nil {
		return nil, err
	}

	return updatedWorld, nil
}

func (cs *worldService) GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error) {
	return cs.worldRepository.GetWorldsList(offset, limit, filter)
}

func (cs *worldService) ClearDatabase() error {
	return cs.worldRepository.ClearDatabase()
}
