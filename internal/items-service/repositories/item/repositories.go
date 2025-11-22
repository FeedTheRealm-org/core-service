package item

import (
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/google/uuid"
)

// ItemRepository defines the interface for item-related database operations.
type ItemRepository interface {
	// CreateItem creates a new item in the database.
	CreateItem(item *models.Item) error

	// CreateItems creates multiple items in the database.
	CreateItems(items []models.Item) error

	// GetItemById retrieves an item by its ID.
	GetItemById(id uuid.UUID) (*models.Item, error)

	// GetAllItems retrieves all items from the database.
	GetAllItems() ([]models.Item, error)

	// GetItemsByCategory retrieves all items of a specific category.
	GetItemsByCategory(category string) ([]models.Item, error)

	// UpdateItem updates an existing item.
	UpdateItem(item *models.Item) error

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error
}

// ItemSpriteRepository defines the interface for item sprite-related database operations.
type ItemSpriteRepository interface {
	// CreateSprite creates a new sprite in the database.
	CreateSprite(sprite *models.ItemSprite) error

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.ItemSprite, error)

	// GetSpritesByCategory retrieves all sprites of a specific category.
	GetSpritesByCategory(category string) ([]models.ItemSprite, error)

	// DeleteSprite deletes a sprite by its ID.
	DeleteSprite(id uuid.UUID) error
}
