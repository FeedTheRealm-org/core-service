package zones

import (
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
)

// ZonesService defines the interface for zone-related operations.
type ZonesService interface {
	// GetWorld retrieves a world by ID.
	GetWorld(worldID uuid.UUID) (*models.WorldData, error)

	// PublishZone creates or updates zone data without starting orchestration.
	PublishZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error)

	// ActivateZone starts orchestration for a zone and marks it active.
	ActivateZone(worldID uuid.UUID, zoneID int) error

	// DeactivateZone stops orchestration for a zone and marks it inactive.
	DeactivateZone(worldID uuid.UUID, zoneID int) error

	// GetWorldZones returns zones for a world.
	GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error)

	// GetWorldZone returns one zone for a world.
	GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error)

	// UpdateZoneStatus updates the status of a zone.
	UpdateZoneStatus(worldID uuid.UUID, zoneID int, isOnline bool) error

	// UpdateZonePlayerCount updates active player count and average player time for a zone.
	UpdateZonePlayerCount(worldID uuid.UUID, zoneID int, activePlayers int, averagePlayerTime int) error

	// GetWorldZonePlayerCounts returns player counts for a world with historic max values.
	GetWorldZonePlayerCounts(worldID uuid.UUID) (int, int, int, int, error)

	// GetAllWorldZonePlayerCounts returns player counts for all worlds with historic max values.
	GetAllWorldZonePlayerCounts() (int, int, int, int, error)

	// StopAllZonesForUser stops all active zones for a specific user.
	StopAllZonesForUser(userID uuid.UUID) error
}
