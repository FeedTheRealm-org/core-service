package item

import (
	"mime/multipart"

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

	// GetItemsByCategory retrieves all items of a specific category.
	GetItemsByCategory(category string) ([]models.Item, error)

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error
}

// ItemSpriteService defines the interface for item sprite-related business logic operations.
type ItemSpriteService interface {
	// UploadSprite uploads a sprite file and saves its metadata.
	UploadSprite(category string, file *multipart.FileHeader) (*models.ItemSprite, error)

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.ItemSprite, error)

	// GetSpritesByCategory retrieves all sprites of a specific category.
	GetSpritesByCategory(category string) ([]models.ItemSprite, error)

	// GetSpriteFile retrieves the file path for a sprite.
	GetSpriteFile(id uuid.UUID) (string, error)

	// DeleteSprite deletes a sprite by its ID.
	DeleteSprite(id uuid.UUID) error
}
