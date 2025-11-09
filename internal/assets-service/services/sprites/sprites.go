package sprites

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/sprites"
	"github.com/google/uuid"
)

type spritesService struct {
	conf              *config.Config
	spritesRepository sprites.SpritesRepository
}

// NewSpritesService creates a new instance of SpritesService.
func NewSpritesService(conf *config.Config, spritesRepository sprites.SpritesRepository) SpritesService {
	return &spritesService{
		conf:              conf,
		spritesRepository: spritesRepository,
	}
}

func (ss *spritesService) GetCategoriesList() ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.Nil}, nil
}

func (ss *spritesService) GetSpritesListByCategory(category uuid.UUID) ([]uuid.UUID, error) {
	return []uuid.UUID{uuid.Nil}, nil
}

func (ss *spritesService) GetSpriteUrl(spriteId uuid.UUID) (string, error) {
	return "", nil
}

func (ss *spritesService) AddCategory(category string) (*models.Category, error) {
	return &models.Category{}, nil
}

func (ss *spritesService) UploadSpriteData(category string, spriteData []byte) (*models.Sprite, error) {
	return &models.Sprite{}, nil
}
