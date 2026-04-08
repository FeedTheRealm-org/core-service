package world

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

// This function checks if the zone is already published,
// if not it checks with payment service
// if the user has available slots to publish a new zone, and if so it allows publishing.
func (cs *worldService) CheckAvaliableZonesForPublish(worldId uuid.UUID, zoneId int) error {
	_, err := cs.worldRepository.GetWorldZone(worldId, zoneId)
	if err != nil {
		userId, wErr := cs.worldRepository.GetUserIdByWorldId(worldId)
		if wErr != nil {
			return wErr
		}

		url := fmt.Sprintf("http://127.0.0.1:%d/internal/subscriptions/users/%s/availability", cs.conf.Server.Port, userId)
		resp, httpErr := http.Get(url)
		if httpErr != nil || resp.StatusCode != http.StatusOK {
			return errors.New("failed to reach payment service to verify slots")
		}
		defer func() {
			_ = resp.Body.Close()
		}()

		var slotsData struct {
			Allowed    bool  `json:"allowed"`
			TotalSlots int64 `json:"total_slots"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&slotsData); err != nil {
			return errors.New("failed to decode slots response")
		}

		if !slotsData.Allowed {
			return errors.New("forbidden: active slots subscription required")
		}

		usedZonesCount, _ := cs.worldRepository.GetTotalZonesCountByUserId(userId)
		if usedZonesCount >= slotsData.TotalSlots {
			return errors.New("forbidden: you have reached the maximum allowed slots for your subscription")
		}
	}

	return nil
}

func (cs *worldService) PublishZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	// if err := cs.CheckAvaliableZonesForPublish(worldID, zoneID); err != nil {
	// 	return nil, err
	// }

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

func (cs *worldService) GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	return cs.worldRepository.GetWorldZone(worldID, zoneID)
}

func (cs *worldService) ClearDatabase() error {
	return cs.worldRepository.ClearDatabase()
}
