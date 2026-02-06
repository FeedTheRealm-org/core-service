package items

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemSpritesService defines the interface for item sprite-related business logic operations.
type ItemSpritesService interface {
	// UploadSprites uploads or overwrites multiple sprite files with provided IDs for a given world and saves their metadata.
	// The ids/files must be paired as id_N/sprite_N from the form. Existing sprites with the same ID will be overwritten.
	UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.Item, error)

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.Item, error)

	// GetAllSprites retrieves all item sprites.
	GetAllSprites() ([]models.Item, error)

	// GetSpriteFile retrieves the file path for a sprite.
	GetSpriteFile(id uuid.UUID) (string, error)

	// DeleteSprite deletes a sprite by its ID and removes the file.
	DeleteSprite(id uuid.UUID) error
}
