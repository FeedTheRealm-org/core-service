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
	ActivateZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error)

	// DeactivateZone stops orchestration for a zone and marks it inactive.
	DeactivateZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error)

	// GetWorldZones returns zones for a world.
	GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error)

	// GetWorldZone returns one zone for a world.
	GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error)
}
