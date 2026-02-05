package sprites

import (
	"fmt"
	"mime/multipart"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/bucket"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/sprites"
	"github.com/google/uuid"
)

type spritesService struct {
	conf *config.Config

	spritesRepository sprites.SpritesRepository
	bucketRepo        bucket.BucketRepository
}

// NewSpritesService creates a new instance of SpritesService.
func NewSpritesService(conf *config.Config, spritesRepository sprites.SpritesRepository, bucketRepo bucket.BucketRepository) SpritesService {
	return &spritesService{
		conf:              conf,
		spritesRepository: spritesRepository,
		bucketRepo:        bucketRepo,
	}
}

func (ss *spritesService) GetCategoriesList() ([]*models.Category, error) {
	return ss.spritesRepository.GetCategoriesList()
}

func (ss *spritesService) GetSpritesListByCategory(category uuid.UUID) ([]*models.Sprite, error) {
	return ss.spritesRepository.GetSpritesListByCategory(category)
}

func (ss *spritesService) GetSpriteUrl(spriteId uuid.UUID) (string, error) {
	sprite, err := ss.spritesRepository.GetSpriteById(spriteId)
	if err != nil {
		return "", err
	}
	return sprite.Url, nil
}

func (ss *spritesService) AddCategory(category string) (*models.Category, error) {
	return ss.spritesRepository.AddCategory(category)
}

func (ss *spritesService) UploadSpriteData(category uuid.UUID, spriteData multipart.File, ext string) (*models.Sprite, error) {
	spriteUniqueUrl := uuid.New().String()

	filePath := fmt.Sprintf("/characters/%s%s", spriteUniqueUrl, ext)
	if err := ss.bucketRepo.UploadFile(filePath, "image/png", spriteData); err != nil {
		return nil, err
	}

	sprite := &models.Sprite{
		Url: filePath,
	}
	if err := ss.spritesRepository.CreateSprite(category, sprite); err != nil {
		return nil, err
	}

	return sprite, nil
}
