package zones

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	world_repository "github.com/FeedTheRealm-org/core-service/internal/world-service/repositories/world"
	"github.com/FeedTheRealm-org/core-service/internal/world-service/services/server_registry"
	"github.com/google/uuid"
)

type zonesService struct {
	conf                  *config.Config
	worldRepository       world_repository.WorldRepository
	serverRegistryService server_registry.ServerRegistryService
}

func NewZonesService(
	conf *config.Config,
	worldRepository world_repository.WorldRepository,
	serverRegistryService server_registry.ServerRegistryService,
) ZonesService {
	return &zonesService{
		conf:                  conf,
		worldRepository:       worldRepository,
		serverRegistryService: serverRegistryService,
	}
}

func (zs *zonesService) GetWorld(worldID uuid.UUID) (*models.WorldData, error) {
	return zs.worldRepository.GetWorldData(worldID)
}

func (zs *zonesService) PublishZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error) {
	return zs.worldRepository.UpsertWorldZone(worldID, zoneID, zoneData)
}

func (zs *zonesService) ActivateZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	zone, err := zs.worldRepository.GetWorldZone(worldID, zoneID)
	if err != nil {
		return nil, err
	}

	if zone.IsActive {
		return zone, nil
	}

	if zs.conf.Server.SubscriptionOn {
		if err := zs.checkAvailableZonesForActivation(worldID); err != nil {
			return nil, err
		}
	}

	if err := zs.serverRegistryService.StartNewJob(worldID, zoneID); err != nil {
		return nil, err
	}

	zone, err = zs.worldRepository.SetWorldZoneActiveState(worldID, zoneID, true)
	if err != nil {
		_ = zs.serverRegistryService.StopJob(worldID, zoneID)
		return nil, err
	}

	if zs.conf.Server.SubscriptionOn {
		userID, err := zs.worldRepository.GetUserIdByWorldId(worldID)
		if err != nil {
			_ = zs.serverRegistryService.StopJob(worldID, zoneID)
			_, _ = zs.worldRepository.SetWorldZoneActiveState(worldID, zoneID, false)
			return nil, err
		}

		if err := zs.updateUsedSlots(userID, 1, true); err != nil {
			_ = zs.serverRegistryService.StopJob(worldID, zoneID)
			_, _ = zs.worldRepository.SetWorldZoneActiveState(worldID, zoneID, false)
			return nil, err
		}
	}

	return zone, nil
}

func (zs *zonesService) DeactivateZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	zone, err := zs.worldRepository.GetWorldZone(worldID, zoneID)
	if err != nil {
		return nil, err
	}

	if !zone.IsActive {
		return zone, nil
	}

	if err := zs.serverRegistryService.StopJob(worldID, zoneID); err != nil {
		return nil, err
	}

	zone, err = zs.worldRepository.SetWorldZoneActiveState(worldID, zoneID, false)
	if err != nil {
		return nil, err
	}

	if zs.conf.Server.SubscriptionOn {
		userID, err := zs.worldRepository.GetUserIdByWorldId(worldID)
		if err != nil {
			return nil, err
		}

		if err := zs.updateUsedSlots(userID, 1, false); err != nil {
			return nil, err
		}
	}

	return zone, nil
}

func (zs *zonesService) GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error) {
	return zs.worldRepository.GetWorldZones(worldID)
}

func (zs *zonesService) GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error) {
	return zs.worldRepository.GetWorldZone(worldID, zoneID)
}

func (zs *zonesService) checkAvailableZonesForActivation(worldID uuid.UUID) error {
	userID, err := zs.worldRepository.GetUserIdByWorldId(worldID)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/subscriptions/internal/users/%s/status", zs.conf.Server.Port, userID)
	resp, err := http.Get(url)
	if err != nil {
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
		return fmt.Errorf("forbidden: you have %d free zones available. Please upgrade your subscription to activate more zones", slotsData.Data.FreeSlots)
	}

	return nil
}

func (zs *zonesService) updateUsedSlots(userID uuid.UUID, numberOfSlots int, areUsed bool) error {
	payload := map[string]interface{}{"slots": numberOfSlots, "are_used": areUsed}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://127.0.0.1:%d/subscriptions/internal/users/%s/used-slots", zs.conf.Server.Port, userID)
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
