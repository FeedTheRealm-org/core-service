package item

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/items-service/repositories/item"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type itemSpriteService struct {
	conf             *config.Config
	spriteRepository item.ItemSpriteRepository
}

// NewItemSpriteService creates a new instance of ItemSpriteService.
func NewItemSpriteService(conf *config.Config, spriteRepository item.ItemSpriteRepository) ItemSpriteService {
	return &itemSpriteService{
		conf:             conf,
		spriteRepository: spriteRepository,
	}
}

func (iss *itemSpriteService) UploadSprite(category string, fileHeader *multipart.FileHeader) (*models.ItemSprite, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	spriteUniqueUrl := uuid.New().String()
	filename := fmt.Sprintf("%s%s", spriteUniqueUrl, ext)

	// Create directory path: ./bucket/items/{category}/
	dirPath := filepath.Join("./bucket/items", category)
	filePath := filepath.Join(dirPath, filename)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}

	// Create destination file
	destFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer destFile.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(destFile, file); err != nil {
		return nil, err
	}

	// Save sprite metadata to database
	sprite := &models.ItemSprite{
		Category: category,
		Url:      filePath,
	}
	if err := iss.spriteRepository.CreateSprite(sprite); err != nil {
		// Clean up the file if database insertion fails
		os.Remove(filePath)
		return nil, err
	}

	logger.Logger.Infof("Item sprite uploaded: %s (ID: %s)", filename, sprite.Id)
	return sprite, nil
}

func (iss *itemSpriteService) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	sprite, err := iss.spriteRepository.GetSpriteById(id)
	if err != nil {
		return nil, err
	}
	return sprite, nil
}

func (iss *itemSpriteService) GetSpritesByCategory(category string) ([]models.ItemSprite, error) {
	sprites, err := iss.spriteRepository.GetSpritesByCategory(category)
	if err != nil {
		return nil, err
	}
	return sprites, nil
}

func (iss *itemSpriteService) GetSpriteFile(id uuid.UUID) (string, error) {
	sprite, err := iss.spriteRepository.GetSpriteById(id)
	if err != nil {
		return "", err
	}
	return sprite.Url, nil
}

func (iss *itemSpriteService) DeleteSprite(id uuid.UUID) error {
	if err := iss.spriteRepository.DeleteSprite(id); err != nil {
		return err
	}
	logger.Logger.Infof("Item sprite deleted: %s", id)
	return nil
}
