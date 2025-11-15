package sprites

import (
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

// SpritesRepository defines the interface for sprites-related database operations.
type SpritesRepository interface {
	GetCategoriesList() ([]*models.Category, error)

	GetSpritesListByCategory(category uuid.UUID) ([]*models.Sprite, error)

	GetSpriteById(spriteId uuid.UUID) (*models.Sprite, error)

	AddCategory(category string) (*models.Category, error)

	CreateSprite(category uuid.UUID, sprite *models.Sprite) error
}
