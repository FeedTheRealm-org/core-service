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

	// GetItemsByCategory retrieves all items of a specific category.
	GetItemsByCategory(categoryId uuid.UUID) ([]models.Item, error)

	// DeleteItem deletes an item by its ID.
	DeleteItem(id uuid.UUID) error

	// ClearAllItems deletes all items (for testing only).
	ClearAllItems() error
}

// ItemCategoryService defines the interface for item category-related business logic operations.
type ItemCategoryService interface {
	// CreateCategory creates a new item category.
	CreateCategory(name string) (*models.ItemCategory, error)

	// GetCategoryById retrieves a category by its ID.
	GetCategoryById(id uuid.UUID) (*models.ItemCategory, error)

	// GetAllCategories retrieves all categories.
	GetAllCategories() ([]models.ItemCategory, error)

	// DeleteCategory deletes a category by its ID.
	DeleteCategory(id uuid.UUID) error

	// ClearAllCategories deletes all categories (for testing only).
	ClearAllCategories() error
}
