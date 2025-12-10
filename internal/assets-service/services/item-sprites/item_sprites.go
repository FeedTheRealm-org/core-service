package itemsprites

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/item-sprites"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
)

type itemSpritesService struct {
	conf       *config.Config
	repository itemsprites.ItemSpritesRepository
}

// NewItemSpritesService creates a new instance of ItemSpritesService.
func NewItemSpritesService(conf *config.Config, repository itemsprites.ItemSpritesRepository) ItemSpritesService {
	return &itemSpritesService{
		conf:       conf,
		repository: repository,
	}
}

func (iss *itemSpritesService) UploadSprite(categoryId uuid.UUID, fileHeader *multipart.FileHeader) (*models.ItemSprite, error) {
	// Validate category exists and get category name for directory
	category, err := iss.repository.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

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

	// Create directory path: ./bucket/sprites/items/{category_name}/
	dirPath := filepath.Join("./bucket/sprites/items", category.Name)
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
		CategoryId: categoryId,
		Url:        filePath,
	}
	if err := iss.repository.CreateSprite(sprite); err != nil {
		// Clean up the file if database insertion fails
		os.Remove(filePath)
		return nil, err
	}

	logger.Logger.Infof("Item sprite uploaded: %s (ID: %s, Category: %s)",
		filename, sprite.Id, category.Name)
	return sprite, nil
}

func (iss *itemSpritesService) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	return iss.repository.GetSpriteById(id)
}

func (iss *itemSpritesService) GetAllSprites() ([]models.ItemSprite, error) {
	return iss.repository.GetAllSprites()
}

func (iss *itemSpritesService) GetSpritesByCategory(categoryId uuid.UUID) ([]models.ItemSprite, error) {
	// Validate category exists
	_, err := iss.repository.GetCategoryById(categoryId)
	if err != nil {
		return nil, err
	}

	return iss.repository.GetSpritesByCategory(categoryId)
}

func (iss *itemSpritesService) GetSpriteFile(id uuid.UUID) (string, error) {
	sprite, err := iss.repository.GetSpriteById(id)
	if err != nil {
		return "", err
	}
	return sprite.Url, nil
}

func (iss *itemSpritesService) DeleteSprite(id uuid.UUID) error {
	// Get sprite to get file path
	sprite, err := iss.repository.GetSpriteById(id)
	if err != nil {
		return err
	}

	// Delete from database
	if err := iss.repository.DeleteSprite(id); err != nil {
		return err
	}

	// Delete file from disk
	if err := os.Remove(sprite.Url); err != nil {
		logger.Logger.Warnf("Failed to delete sprite file %s: %v", sprite.Url, err)
		// Don't fail the request - database record is already deleted
	}

	logger.Logger.Infof("Item sprite deleted: %s", id)
	return nil
}

func (iss *itemSpritesService) GetAllCategories() ([]models.ItemCategory, error) {
	return iss.repository.GetAllCategories()
}
