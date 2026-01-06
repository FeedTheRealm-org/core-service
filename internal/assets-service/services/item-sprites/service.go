package itemsprites

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// ItemSpritesService defines the interface for item sprite-related business logic operations.
type ItemSpritesService interface {
	// UploadSprites uploads multiple sprite files with provided IDs for a given world and saves their metadata.
	UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.ItemSprite, error)

	// GetSpriteById retrieves a sprite by its ID.
	GetSpriteById(id uuid.UUID) (*models.ItemSprite, error)

	// GetAllSprites retrieves all item sprites.
	GetAllSprites() ([]models.ItemSprite, error)

	// GetSpriteFile retrieves the file path for a sprite.
	GetSpriteFile(id uuid.UUID) (string, error)

	// DeleteSprite deletes a sprite by its ID and removes the file.
	DeleteSprite(id uuid.UUID) error
}
