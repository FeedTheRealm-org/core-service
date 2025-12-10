package itemsprites

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemSpritesService defines the interface for item sprite-related business logic operations.
type ItemSpritesService interface {
	// UploadSprite uploads a sprite file and saves its metadata.
	UploadSprite(categoryId uuid.UUID, file *multipart.FileHeader) (*models.ItemSprite, error)

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.ItemSprite, error)

	// GetAllSprites retrieves all item sprites.
	GetAllSprites() ([]models.ItemSprite, error)

	// GetSpritesByCategory retrieves all sprites for a category.
	GetSpritesByCategory(categoryId uuid.UUID) ([]models.ItemSprite, error)

	// GetSpriteFile retrieves the file path for a sprite.
	GetSpriteFile(id uuid.UUID) (string, error)

	// DeleteSprite deletes a sprite by its ID and removes the file.
	DeleteSprite(id uuid.UUID) error

	// GetAllCategories retrieves all item categories (for listing).
	GetAllCategories() ([]models.ItemCategory, error)
}
