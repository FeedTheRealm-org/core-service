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

	// UpdateWorldData updates the data and description for an existing world, only if owned by userId.
	UpdateWorldData(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error)

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error)

	ClearDatabase() error
}
