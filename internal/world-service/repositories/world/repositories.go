package world

import (
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
)

// WorldRepository defines the interface for world-related database operations.
type WorldRepository interface {
	// StoreWorldData handles the storing of world information.
	StoreWorldData(newWorldData *models.WorldData) error

	// GetWorldData retrieves information for a specific world.
	GetWorldData(worldID uuid.UUID) (*models.WorldData, error)

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(offset int, limit int) ([]*models.WorldData, error)
}
