package world

import (
	"errors"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/google/uuid"
	"gorm.io/datatypes"
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
	if len(newWorldData.CreateableData) == 0 {
		newWorldData.CreateableData = datatypes.JSON([]byte("{}"))
	}

	createdWorld, err := cs.worldRepository.StoreWorldData(newWorldData)
	if err != nil {
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

	return updatedWorld, nil
}

func (cs *worldService) UpdateCreateableData(worldID uuid.UUID, userId uuid.UUID, createableData []byte) (*models.WorldData, error) {
	return cs.worldRepository.UpdateCreateableData(worldID, userId, createableData)
}

func (cs *worldService) PublishZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	zone, err := cs.worldRepository.UpsertWorldZone(worldID, zoneID, zoneData)
	if err != nil {
		return nil, err
	}

	if err := cs.serverRegistryService.StartNewJob(worldID, zoneID); err != nil {
		return nil, err
	}

	return zone, nil
}

func (cs *worldService) DeleteWorld(worldID uuid.UUID, userId uuid.UUID) error {
	worldData, err := cs.worldRepository.GetWorldData(worldID)
	if err != nil {
		return err
	}
	if worldData.UserId != userId {
		return errors.New("forbidden: user does not own this world")
	}

	zones, err := cs.worldRepository.GetWorldZones(worldID)
	if err != nil {
		return err
	}

	if err := cs.worldRepository.DeleteWorldData(worldID); err != nil {
		return err
	}

	for _, zone := range zones {
		if err := cs.serverRegistryService.StopJob(worldID, zone.ID); err != nil {
			return err
		}
	}

	return nil
}

func (cs *worldService) GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error) {
	return cs.worldRepository.GetWorldsList(offset, limit, filter)
}

func (cs *worldService) GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error) {
	return cs.worldRepository.GetWorldZones(worldID)
}

func (cs *worldService) ClearDatabase() error {
	return cs.worldRepository.ClearDatabase()
}
