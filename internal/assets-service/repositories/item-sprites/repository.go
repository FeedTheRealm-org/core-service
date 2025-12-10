package itemsprites

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemSpritesRepository defines the interface for item sprites-related database operations.
type ItemSpritesRepository interface {
	// CreateSprite creates a new item sprite in the database.
	CreateSprite(sprite *models.ItemSprite) error

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.ItemSprite, error)

	// GetAllSprites retrieves all item sprites.
	GetAllSprites() ([]models.ItemSprite, error)

	// GetSpritesByCategory retrieves all sprites for a specific category.
	GetSpritesByCategory(categoryId uuid.UUID) ([]models.ItemSprite, error)

	// DeleteSprite deletes a sprite by its ID.
	DeleteSprite(id uuid.UUID) error

	// Category validation methods (reads from items-service table)
	// GetCategoryById retrieves a category by ID for validation.
	GetCategoryById(id uuid.UUID) (*models.ItemCategory, error)

	// GetAllCategories retrieves all item categories.
	GetAllCategories() ([]models.ItemCategory, error)
}
