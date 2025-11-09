package sprites

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// SpritesRepository defines the interface for sprites-related database operations.
type SpritesRepository interface {
	GetCategoriesList() ([]uuid.UUID, error)

	GetSpritesListByCategory(category uuid.UUID) ([]uuid.UUID, error)

	GetSpriteById(spriteId uuid.UUID) (*models.Sprite, error)

	AddCategory(category string) (*models.Category, error)

	UploadSpriteData(category uuid.UUID, spriteData []byte) (*models.Sprite, error)
}
