package itemsprites

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/assets-service/models"
	itemsprites "github.com/FeedTheRealm-org/core-service/internal/assets-service/repositories/item-sprites"
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

func (iss *itemSpritesService) UploadSprites(worldID uuid.UUID, ids []uuid.UUID, files []*multipart.FileHeader) ([]*models.ItemSprite, error) {
	if len(ids) != len(files) {
		return nil, fmt.Errorf("number of ids and files must match")
	}
	var result []*models.ItemSprite
	for i, fileHeader := range files {
		id := ids[i]
		file, err := fileHeader.Open()
		if err != nil {
			return nil, err
		}
		defer file.Close()

		ext := filepath.Ext(fileHeader.Filename)
		filename := fmt.Sprintf("%s%s", id.String(), ext)
		dirPath := filepath.Join("bucket", "worlds", worldID.String(), "items")
		filePath := filepath.Join(dirPath, filename)
		if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
			return nil, err
		}
		destFile, err := os.Create(filePath)
		if err != nil {
			return nil, err
		}
		if _, err := io.Copy(destFile, file); err != nil {
			destFile.Close()
			return nil, err
		}
		destFile.Close()

		sprite := &models.ItemSprite{
			Id:  id,
			Url: filePath,
		}
		if err := iss.repository.CreateSprite(sprite); err != nil {
			_ = os.Remove(filePath)
			return nil, err
		}
		logger.Logger.Infof("Item sprite uploaded: %s (ID: %s)", filename, sprite.Id)
		result = append(result, sprite)
	}
	return result, nil
}

func (iss *itemSpritesService) GetSpriteById(id uuid.UUID) (*models.ItemSprite, error) {
	return iss.repository.GetSpriteById(id)
}

func (iss *itemSpritesService) GetAllSprites() ([]models.ItemSprite, error) {
	return iss.repository.GetAllSprites()
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

// (Item sprite categories removed)
