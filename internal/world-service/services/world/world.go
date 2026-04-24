package world

import (
	"bytes"
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

func (cs *worldService) UpdateUsedSlots(userId uuid.UUID, numberOfSlots int, areUsed bool) error {
	payload := map[string]interface{}{"slots": numberOfSlots, "are_used": areUsed}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/subscriptions/internal/users/%s/used-slots", cs.conf.Server.Port, userId)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("failed to reach payment service to update used slots")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update used slots, payment service returned status: %d", resp.StatusCode)
	}

	return nil
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

	activeZonesCount := 0
	for _, zone := range zones {
		if !zone.IsActive {
			continue
		}

		if err := cs.serverRegistryService.StopJob(worldID, zone.ID); err != nil {
			return err
		}

		activeZonesCount++
	}

	if cs.conf.Server.SubscriptionOn && activeZonesCount > 0 {
		if err := cs.UpdateUsedSlots(userId, activeZonesCount, false); err != nil {
			return err
		}
	}

	return nil
}

func (cs *worldService) GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error) {
	return cs.worldRepository.GetWorldsList(offset, limit, filter)
}

func (cs *worldService) ClearDatabase() error {
	return cs.worldRepository.ClearDatabase()
}
