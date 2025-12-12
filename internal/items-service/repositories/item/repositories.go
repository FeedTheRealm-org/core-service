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

	// UpdateItem updates an existing item.
	UpdateItem(item *models.Item) error

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error

	// DeleteAll deletes all items (for testing only).
	DeleteAll() error
}
