package items

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemService defines the interface for item-related business logic operations.
type ItemService interface {
	// UploadSprite uploads or overwrites a single sprite file with the provided ID for a given world and saves its metadata.
	// idStr is the raw ID value from the client (string). The service will validate/parse it as a UUID.
	UploadSprite(worldID uuid.UUID, categoryId uuid.UUID, id uuid.UUID, file *multipart.FileHeader) (*models.Item, error)

	// GetItemById retrieves an item by its ID.
	GetItemById(id uuid.UUID) (*models.Item, error)

	GetItemsListByCategory(worldId uuid.UUID, categoryId uuid.UUID) ([]*models.Item, error)

	// GetAllItems retrieves all items.
	GetAllItems() ([]*models.Item, error)

	AddCategory(name string) (*models.ItemCategory, error)
}
