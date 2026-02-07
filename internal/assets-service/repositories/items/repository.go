package items

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemRepository defines the interface for item-related database operations.
type ItemRepository interface {
	// UpsertItem inserts or updates an item in the database.
	UpsertItem(item *models.Item) error

	// GetItemById retrieves an item by its ID.
	GetItemById(id uuid.UUID) (*models.Item, error)

	// GetAllItems retrieves all items.
	GetAllItems() ([]*models.Item, error)

	GetItemsListByCategory(worldId uuid.UUID, categoryId uuid.UUID) ([]*models.Item, error)

	// DeleteSprite deletes a sprite by its ID.
	DeleteSprite(id uuid.UUID) error

	AddCategory(name string) (*models.ItemCategory, error)
}
