package items

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemService defines the interface for item-related business logic operations.
type ItemService interface {
	// UploadSprites uploads or overwrites multiple sprite files with provided IDs for a given world and saves their metadata.
	// The ids/files must be paired as id_N/sprite_N from the form. Existing sprites with the same ID will be overwritten.
	UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.Item, error)

	// GetItemById retrieves an item by its ID.
	GetItemById(id uuid.UUID) (*models.Item, error)

	GetItemsListByCategory(categoryId uuid.UUID) ([]*models.Item, error)

	// GetAllItems retrieves all items.
	GetAllItems() ([]*models.Item, error)

	AddCategory(name string) (*models.ItemCategory, error)
}
