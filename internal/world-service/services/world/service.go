package world

import (
	"github.com/FeedTheRealm-org/core-service/internal/world-service/models"
	"github.com/google/uuid"
)

// WorldService defines the interface for character-related operations.
type WorldService interface {
	// PublishWorld handles the publishing of world information.
	PublishWorld(newWorldData *models.WorldData) (*models.WorldData, error)

	// GetWorldData retrieves information for a specific world.
	GetWorld(worldID uuid.UUID) (*models.WorldData, error)

	// UpdateWorld updates the data and description for an existing world, only if owned by userId.
	UpdateWorld(worldID uuid.UUID, userId uuid.UUID, data []byte, description string) (*models.WorldData, error)

	// GetWorldsList retrieves a paginated list of worlds.
	GetWorldsList(offset int, limit int, filter string) ([]*models.WorldData, error)

	ClearDatabase() error
}
