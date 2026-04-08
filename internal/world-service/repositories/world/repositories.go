package world

import (
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
)

// WorldRepository defines the interface for world-related database operations.
type WorldRepository interface {
	// StoreWorldData handles the storing of world information.
	StoreWorldData(newWorldData *models.WorldData) (*models.WorldData, error)

	// GetWorldData retrieves information for a specific world.
	GetWorldData(worldID uuid.UUID) (*models.WorldData, error)

	// UpdateWorldData updates the data, description and zone for an existing world, only if owned by userId.
	UpdateWorldData(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error)

	// UpdateCreateableData updates createable data for an existing world, only if owned by userId.
	UpdateCreateableData(worldID uuid.UUID, userId uuid.UUID, createableData []byte) (*models.WorldData, error)

	// UpsertWorldZone creates or updates a zone for a world.
	UpsertWorldZone(worldID uuid.UUID, zoneID int, zoneData []byte) (*models.WorldZone, error)

	// DeleteWorldData deletes a world from the database, only if owned by userId.
	DeleteWorldData(worldID uuid.UUID) error

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error)

	// GetWorldZones retrieves available zones for a specific world.
	GetWorldZones(worldID uuid.UUID) ([]*models.WorldZone, error)

	// GetWorldZone retrieves a specific zone for a world.
	GetWorldZone(worldID uuid.UUID, zoneID int) (*models.WorldZone, error)

	GetUserIdByWorldId(worldID uuid.UUID) (uuid.UUID, error)

	// GetUserWorlds retrieves all worlds owned by a specific user.
	GetTotalZonesCountByUserId(userId uuid.UUID) (int64, error)

	ClearDatabase() error
}
