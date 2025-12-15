package item

import (
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/google/uuid"
)

// ItemService defines the interface for item-related business logic operations.
type ItemService interface {
	// CreateItem creates a new item.
	CreateItem(item *models.Item) error

	// CreateItems creates multiple items at once.
	CreateItems(items []models.Item) error

	// GetItemById retrieves an item by its ID.
	GetItemById(id uuid.UUID) (*models.Item, error)

	// GetAllItems retrieves all items.
	GetAllItems() ([]models.Item, error)

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error

	// ClearAllItems deletes all items (for testing only).
	ClearAllItems() error

	// UpdateItemSprite updates the sprite associated to an item.
	UpdateItemSprite(id uuid.UUID, spriteId uuid.UUID) error
}
