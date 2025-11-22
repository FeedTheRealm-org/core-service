package sprites

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

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

	filename := fmt.Sprintf("%s%s", spriteUniqueUrl, ext)
	dirPath := "./bucket/sprites"
	filePath := filepath.Join(dirPath, filename)

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}

	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = destFile.Close()
	}()

	if _, err := io.Copy(destFile, spriteData); err != nil {
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
