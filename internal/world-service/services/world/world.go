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

func (cs *worldService) CheckAvaliableZonesForPublish(worldId uuid.UUID, zoneId int) error {
	userId, wErr := cs.worldRepository.GetUserIdByWorldId(worldId)
	if wErr != nil {
		return wErr
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/payments/subscriptions/internal/users/%s/status", cs.conf.Server.Port, userId)
	resp, httpErr := http.Get(url)
	if httpErr != nil {
		return errors.New("failed to reach payment service to verify slots")
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusNotFound {
		return errors.New("forbidden: active slots subscription required")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reach payment service to verify slots: status %d", resp.StatusCode)
	}

	var slotsData struct {
		Data struct {
			Allowed   bool `json:"allowed"`
			FreeSlots int  `json:"free_slots"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&slotsData); err != nil {
		return errors.New("failed to decode slots response")
	}

	if !slotsData.Data.Allowed {
		return errors.New("active slots subscription required")
	}

	if slotsData.Data.FreeSlots <= 0 {
		return fmt.Errorf("forbidden: you have %d free zones available. Please upgrade your subscription to publish more zones", slotsData.Data.FreeSlots)
	}

	return nil
}

func (cs *worldService) UpdateUsedSlots(userId uuid.UUID, numberOfSlots int, areUsed bool) error {
	payload := map[string]interface{}{"slots": numberOfSlots, "are_used": areUsed}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/payments/subscriptions/internal/users/%s/used-slots", cs.conf.Server.Port, userId)
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

func (cs *worldService) PublishZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	isNewZone := false
	_, err := cs.worldRepository.GetWorldZone(worldID, zoneID)
	if err != nil {
		isNewZone = true
	}

	if isNewZone {
		if err := cs.CheckAvaliableZonesForPublish(worldID, zoneID); err != nil {
			return nil, err
		}
	}

	zone, err := cs.worldRepository.UpsertWorldZone(worldID, zoneID, zoneData)
	if err != nil {
		return nil, err
	}

	if err := cs.serverRegistryService.StartNewJob(worldID, zoneID); err != nil {
		return nil, err
	}

	if isNewZone {
		userId, err := cs.worldRepository.GetUserIdByWorldId(worldID)
		if err == nil {
			if err := cs.UpdateUsedSlots(userId, 1, true); err != nil {
				return nil, err
			}
		}
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

	if len(zones) > 0 {
		if err := cs.UpdateUsedSlots(userId, len(zones), false); err != nil {
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
