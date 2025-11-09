package sprites

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/google/uuid"
)

type spritesRepository struct {
	conf *config.Config
	db   *config.DB
}

// NewSpritesRepository creates a new instance of CharacteRepository.
func NewSpritesRepository(conf *config.Config, db *config.DB) SpritesRepository {
	return &spritesRepository{
		conf: conf,
		db:   db,
	}
}

func (sr *spritesRepository) GetCategoriesList() ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.Nil}, nil
}

func (sr *spritesRepository) GetSpritesListByCategory(category uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.Nil}, nil
}

func (sr *spritesRepository) GetSpriteById(spriteId uuid.UUID) (*models.Sprite, error) {
	return &models.Sprite{}, nil
}

func (sr *spritesRepository) AddCategory(category string) (*models.Category, error) {
	return &models.Category{}, nil
}

func (sr *spritesRepository) UploadSpriteData(category uuid.UUID, spriteData []byte) (*models.Sprite, error) {
	return &models.Sprite{}, nil
}
