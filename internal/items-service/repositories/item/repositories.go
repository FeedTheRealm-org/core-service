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
	GetItemsByCategory(categoryId uuid.UUID) ([]models.Item, error)

	// UpdateItem updates an existing item.
	UpdateItem(item *models.Item) error

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error

	// DeleteAll deletes all items (for testing only).
	DeleteAll() error
}

// ItemCategoryRepository defines the interface for item category-related database operations.
type ItemCategoryRepository interface {
	// CreateCategory creates a new item category in the database.
	CreateCategory(category *models.ItemCategory) error

	// GetCategoryById retrieves a category by its ID.
	GetCategoryById(id uuid.UUID) (*models.ItemCategory, error)

	// GetAllCategories retrieves all categories from the database.
	GetAllCategories() ([]models.ItemCategory, error)

	// DeleteCategory deletes a category by its ID.
	DeleteCategory(id uuid.UUID) error

	// CountItemsUsingCategory counts items that reference a specific category.
	CountItemsUsingCategory(categoryId uuid.UUID) (int64, error)

	// CountSpritesUsingCategory counts sprites that reference a specific category.
	CountSpritesUsingCategory(categoryId uuid.UUID) (int64, error)

	// DeleteAll deletes all categories (for testing only).
	DeleteAll() error
}
