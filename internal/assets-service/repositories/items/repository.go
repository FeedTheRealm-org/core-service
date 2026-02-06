package items

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemSpritesRepository defines the interface for item sprites-related database operations.
type ItemSpritesRepository interface {
	// UpsertSprite inserts or updates an item sprite in the database.
	UpsertSprite(sprite *models.Item) error

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.Item, error)

	// GetAllSprites retrieves all item sprites.
	GetAllSprites() ([]models.Item, error)

	// DeleteSprite deletes a sprite by its ID.
	DeleteSprite(id uuid.UUID) error
}
