package cosmetics

import (
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type SpritesService interface {

	// GetCategoriesList retrieves a list of sprite categories.
	GetCategoriesList() ([]*models.Category, error)

	// GetSpritesListByCategory retrieves a list of sprites for a given category.
	GetSpritesListByCategory(category uuid.UUID) ([]*models.Sprite, error)

	// GetSpriteUrl handles the retrieval of sprite file URL.
	GetSpriteUrl(spriteId uuid.UUID) (string, error)

	// AddCategory handles the addition of a new sprite category.
	AddCategory(category string) (*models.Category, error)

	// UploadSpriteData handles the upload of sprite file.
	UploadSpriteData(category uuid.UUID, spriteData multipart.File, ext string) (*models.Sprite, error)
}
